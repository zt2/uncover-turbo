# Plan · 自然语言跨多测绘引擎查询与聚合去重(三层架构)

> SDD 阶段二产出物。基于已确认的 `spec.md` 设计「怎么做」。
> **状态:已确认。** 可进入阶段三（Tasks）。

## 0. 关键前置发现(已核实 uncover v1.2.1 源码)

1. **Execute 是笛卡尔积**:`Service.Execute` 内部 `for q in Queries { for agent in Agents }`。
   → **每个引擎单独构造一个 `uncover.Service`(单 agent + 单查询)并发执行**,规避串扰。
2. **结果流 + 原始字节**:`Execute(ctx)` 返回 `<-chan sources.Result`,每条含 `Source/IP/Port/Host/Url` + 原始 `Raw []byte` → 富字段来源。
3. **引擎凭据走 uncover 自己的 Provider**(env / provider-config)。本项目 YAML 只管 LLM 后端 + 引擎启停 + 去重/输出偏好。

## 1. 核心架构:三层查询链路

```
                 (仅此一步用 LLM)         (确定性,纯代码)
  自然语言 NL ──► 通用中间层 IR ──► 各引擎专有查询语法 ──► uncover 并发查询 ──► 聚合去重 ──► []Asset
                     ▲
        机器调用方可直接从这里注入 IR(绕过 NL / LLM)
```

- **上层(NL→IR)**:LLM 把自然语言翻译成**唯一目标语法——通用 IR**(可校验、可序列化)。LLM 只需掌握一种语法,鲁棒性远高于「为每个引擎各生成一种语法」。
- **中层(IR)**:引擎无关的**机器查询表示**(JSON 可序列化的 AST)。是本工具对**非人调用方**的稳定接口。
- **下层(IR→引擎)**:每个引擎一个**确定性编译器**,把 IR 编译为该引擎专有语法。同一 IR 对同一引擎恒定输出同一查询串,不经 LLM。

这直接兑现 spec 的 **AC1 / AC9**,并大幅削弱「LLM 产出各引擎语法错误」的旧风险(R3)。

## 2. 中间层 IR(pkg/queryir)

### 2.1 AST 结构(JSON 可序列化)

```go
// Expr 是查询 AST 节点,四选一:And / Or / Not / Match。
type Expr struct {
    And   []Expr `json:"and,omitempty"`
    Or    []Expr `json:"or,omitempty"`
    Not   *Expr  `json:"not,omitempty"`
    Match *Match `json:"match,omitempty"`
}
type Match struct {
    Field Field  `json:"field"` // 规范字段名(见 2.2)
    Op    Op     `json:"op"`    // eq | contains | ne
    Value string `json:"value"`
}
```

机器接口示例(「美国开放 3306 的主机」):
```json
{"and":[
  {"match":{"field":"port","op":"eq","value":"3306"}},
  {"match":{"field":"country","op":"eq","value":"US"}}
]}
```

### 2.2 规范字段词汇(v1 核心集,可扩展)

| 规范字段 | 含义 |
|---|---|
| `ip` | IP 地址 |
| `port` | 端口 |
| `domain` | 注册域名 |
| `host` | 主机名 |
| `title` | HTTP 标题 |
| `body` | 响应体 / 正文关键字 |
| `product` | 组件 / 产品名 |
| `country` | 国家码 |
| `org` | 组织 |
| `asn` | ASN |
| `protocol` | 服务 / 协议 |
| `status` | HTTP 状态码 |
| `cert.subject_cn` | 证书 CN |
| `os` | 操作系统 |

> 词汇是**开放集**:新增字段只需在此表 + 各引擎映射表补一行。

### 2.3 能力
- `Validate() error`:结构合法(四选一)、字段在词汇内、op 合法、value 非空。
- `String() string`:人类可读形式(如 `port eq "3306" AND country eq "US"`),供 verbose/展示。
- JSON 编解码即 AST 的机器序列化(供 AC9 机器调用)。

## 3. 下层编译器(pkg/compiler)

```go
var ErrUnsupported = errors.New("field/op not supported by engine")

type Compiler interface {
    Engine() string
    Compile(expr queryir.Expr) (string, error) // 不支持的字段/op → 返回 ErrUnsupported
}
```

- 每个引擎一个编译器 + 一张**字段映射表** `map[queryir.Field]fieldSpec`(引擎字段名 + eq/contains/ne 的格式化方式),再加该引擎的布尔/分组语法。
- v1 覆盖 **fofa / censys / hunter / zoomeye** 四家,映射示例(端口为例):

  | 规范 | fofa | censys | hunter | zoomeye |
  |---|---|---|---|---|
  | `port=3306` | `port="3306"` | `services.port: 3306` | `port="3306"` | `port:"3306"` |
  | 布尔 AND/OR/NOT | `&&` `\|\|` `!` | `and` `or` `not` | `&&` `\|\|` `!` | 空格(隐式 AND)/`-` |

  (完整字段映射表在 Tasks/实现阶段依各引擎官方文档定稿并落成代码常量。)
- **不支持策略(AC9 明确行为)**:某引擎无法表达 IR 中的字段/op → `Compile` 返回 `ErrUnsupported`,编排层将**跳过该引擎**并记入 `EngineErrors`(绝不静默产出错误查询)。「严格跳过 vs 尽力而为」可作为未来配置项,v1 默认严格跳过。

## 4. 上层 LLM 翻译器(pkg/llm + pkg/prompts)

- 接口(便于 mock):
  ```go
  type Translator interface {
      Translate(ctx context.Context, nl string) (queryir.Expr, error) // NL -> IR
  }
  ```
- 实现:go-openai + 自定义 `BaseURL`(兼容 OpenRouter)。**单一 system prompt**:说明 IR 的 JSON schema + 规范字段词汇 + 只准输出 JSON;拿到回复后 `json.Unmarshal` + `Validate`,失败可**重试一次**(把校验错误回喂模型),再失败则报错。
- 相比旧版「每引擎一个 prompt」,现在**只有一个 NL→IR prompt**,LLM 表面大幅收敛。

## 5. 数据模型(pkg/asset)

```go
type Asset struct {
    IP        string                     `json:"ip,omitempty"`
    Port      int                        `json:"port,omitempty"`
    Hosts     []string                   `json:"hosts,omitempty"`
    URLs      []string                   `json:"urls,omitempty"`
    Sources   []string                   `json:"sources"`
    Fields    map[string]any             `json:"fields,omitempty"`     // 各引擎 Raw 扁平并集,首个非空优先
    PerSource map[string]json.RawMessage `json:"per_source,omitempty"` // 各引擎原始 JSON,无损归档
}
```
- **去重主键 IP:Port**(已确认);`IP` 空则退化 `host:Port` 兜底。
- 合并:`Hosts/URLs/Sources` 并集去重排序;`Fields` 逐 key 并入(首个非空优先);`PerSource` 无损保留。

## 6. 核心编排(pkg/search)

```go
type Result struct {
    Assets       []asset.Asset
    IR           queryir.Expr        // 本次使用的 IR(NL 入口时由 LLM 产出;IR 入口时即入参)
    Queries      map[string]string   // engine -> 编译出的查询串(观测/JSON 元信息)
    EngineErrors map[string]error    // 每引擎错误(翻译/编译/查询失败,部分失败不致命)
}

type Service struct{ tr llm.Translator; compilers map[string]compiler.Compiler; cfg *config.Config }

// 机器入口:直接吃 IR。
func (s *Service) SearchIR(ctx context.Context, ir queryir.Expr) (*Result, error)
// 人类入口:NL --LLM--> IR,再走 SearchIR。
func (s *Service) SearchNL(ctx context.Context, nl string) (*Result, error)
```

`SearchIR` 流程:
1. 校验 IR。
2. 对每个 `enabled` 引擎:用其 `Compiler.Compile(ir)` 得查询串(`ErrUnsupported` → 记 EngineErrors 并跳过)。
3. 对有查询串的引擎**各自** `uncover.New(Agents:[engine],Queries:[q]).Execute(ctx)`(规避 §0.1),结果流并发汇入聚合器(带锁),按主键合并进 `map[key]*Asset`。
4. 汇成稳定排序 `[]Asset`,连同 IR/Queries/EngineErrors 返回。

`SearchNL` = `tr.Translate(nl)` → `SearchIR`。→ 中层对机器调用方独立可用(AC9),核心零 I/O(AC7)。

## 7. 输出渲染(pkg/render)

```go
type Renderer interface { Render(w io.Writer, res *search.Result) error }
```
- `jsonRenderer`:每资产一 JSON 对象(sources + 合并字段 + 可选 per_source);可选顶层 `meta`(IR、各引擎 query)。
- `textRenderer`:人类可读行 + 颜色。复用现有 `logrusorgru/aurora`(不新增颜色依赖)。颜色 `auto`(仅 TTY 上色)/`always`/`never`;`-no-color` 等价 never。

## 8. 配置(pkg/config) —— 同前次 Plan

YAML @ `~/.config/uncover-turbo/config.yaml`,优先级 **CLI > 环境变量 > 文件 > 默认**;`llm.{base_url,model,api_key,temperature}`、`engines.enabled`、`output.{format,color}`、`limit`。敏感项可用 `UNCOVER_TURBO_LLM_*` 环境变量覆盖。引擎凭据仍走 uncover。新增依赖 `gopkg.in/yaml.v3`。

## 9. CLI(cmd/uncover-turbo/main.go)

薄封装。flags:`-q`(自然语言)、`-ir`(直接传 IR 的 JSON,与 `-q` 二选一,兑现机器入口)、`-e/-engines`、`-config`、`-json`、`-no-color`、`-l/-limit`、`-proxy`、`-v`、`-delay`。
流程:parse → `config.Load` + flag 覆盖 → 构造 Translator/Compilers/Service → `-ir` 则 `SearchIR` 否则 `SearchNL` → 选 Renderer 输出 → 退出码。`-v` 打印 IR 与各引擎编译出的 query(观测,支撑 AC1/AC9)。

## 10. 包结构总览

```
cmd/uncover-turbo/main.go
pkg/config/          配置载入(YAML + env + 优先级)
pkg/queryir/         IR:AST、词汇、Validate、String、JSON
pkg/compiler/        Compiler 接口 + registry + 数据驱动 dialect + 四引擎映射(单包)
pkg/prompts/         NL→IR 单一 system prompt(里程碑 B)
pkg/llm/             Translator:NL → queryir.Expr(go-openai + BaseURL)(里程碑 B)
pkg/asset/           Asset 模型 + Aggregator
pkg/search/          编排:SearchNL / SearchIR
pkg/render/          JSON / 彩色文本
specs/cross-engine-search/   本 SDD 文档
```
> **实现修订**:各引擎编译器最终采用**单包数据驱动 dialect**(`pkg/compiler` 内 `dialect.go`+`engines.go`),而非每引擎一个子包 —— 引擎差异仅是「字段映射表 + 布尔语法 + 取反模型」的数据,单包更简洁、无样板、同样可表驱动测试。§3 的「每引擎一个编译器」按此理解为「每引擎一个 dialect 值」。
> 旧 `pkg/bot` 被 `pkg/llm`+`pkg/prompts` 取代;旧 `main.go` 完全重写;README 更新纳入 Tasks。

## 11. 测试策略(支撑 AC8,得益于确定性编译更好测)

- `pkg/queryir`:JSON 往返、`Validate`、`String`。
- `pkg/compiler/*`:**表驱动**——给定 IR 断言**精确**引擎查询串(纯确定性、零网络)→ 强覆盖 AC1/AC9 下层与不支持字段跳过。
- `pkg/search`:mock `Translator` 产出固定 IR + 假引擎结果(含 Raw),断言编译分发、不支持跳过、聚合去重、富字段并集、per_source 无损 → 覆盖 AC1/AC2/AC3/AC5/AC9。
- `pkg/config`:YAML + env + 优先级。
- `pkg/render`:固定 Result 断言 JSON 结构与(禁色)文本。
- 门禁:`go build ./... && go vet ./... && go test ./...`。

## 12. 关键决策与取舍

| 决策 | 选择 | 取舍 |
|---|---|---|
| 查询链路 | NL→IR→引擎 三层 | LLM 只学一种语法 + 下层确定性,鲁棒且机器可调;需为每引擎写编译器与字段映射(主要成本) |
| IR 形态 | JSON 可序列化 AST | 机器友好、可校验;人读性靠 `String()` 补 |
| 引擎↔查询 | 每引擎独立 Execute | 规避 uncover 笛卡尔积 |
| 不支持字段 | 严格跳过 + 记错误 | 保正确性(不发错查询);「尽力而为」留作未来配置 |
| 富字段 | Fields 并集 + PerSource 无损 | 便捷视图 + 原始两全 |
| 颜色/依赖 | 复用 aurora;仅新增 yaml.v3 | 控制依赖膨胀 |

## 13. 风险点

- **R1 · v1 字段/引擎覆盖有限**:v1 只覆盖核心字段 × 4 引擎;超出的字段该引擎跳过。以「开放集 + 映射表」保证可增量扩展;`-v` 暴露被跳过引擎与原因。
- **R2 · 各引擎语法细节**:fofa/censys/hunter/zoomeye 语法差异大(尤其 censys 的 `services.*` 与其它家差别显著);映射表需对官方文档,实现阶段逐引擎核对,表驱动测试兜底。
- **R3 · LLM 产出非法 IR**:以「JSON schema 约束 + Validate + 一次重试」应对;失败明确报错而非静默。
- **R4 · Raw 结构差异**:`Fields` 扁平并集语义不完全统一,以 `PerSource` 无损保底,文档说明。
- **R5 · 破坏性变更**:CLI 与旧版不兼容,README/示例同步更新(Tasks 覆盖)。

---

### 后续
确认本 Plan 后进入 **阶段三 · Tasks**,拆为有序、可独立验证的任务清单 `specs/cross-engine-search/tasks.md`。
