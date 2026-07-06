# uncover-turbo

一个自然语言驱动的**跨多测绘引擎**查询与聚合工具。输入一句自然语言,它会翻译成通用查询、并发查询多个测绘搜索引擎(FOFA / 360 Quake / Censys / ZoomEye / Hunter 等),再把结果**聚合去重**成统一的资产视图。

项目基于 project-discovery 的 [uncover](https://github.com/projectdiscovery/uncover) 改造而成。

## 核心架构:三层查询链路

```
                 (仅此一步用 LLM)          (确定性,纯代码)
  自然语言 NL ──► 通用中间层 IR ──► 各引擎专有查询语法 ──► uncover 并发查询 ──► 聚合去重 ──► 资产
                     ▲
        机器调用方可直接从这里注入 IR(绕过自然语言 / LLM)
```

- **上层(NL → IR)**:LLM 把自然语言翻译成**唯一目标语法——通用 IR**。LLM 只需掌握一种语法,鲁棒性远高于「为每个引擎各生成一种语法」。
- **中层(IR)**:引擎无关的**机器查询表示**(JSON 可序列化的布尔 AST)。非人程序可**直接构造 IR**,绕过 LLM,得到同样的跨引擎聚合结果。
- **下层(IR → 引擎)**:每个引擎一个**确定性编译器**,把 IR 编译为该引擎专有语法。同一 IR 恒定产出同一查询串,不经 LLM。

好处:程序化调用友好;各引擎语法确定性生成,基本消除「LLM 产出各引擎语法错误」的老问题。

## 特性

- **跨引擎并发**查询 + **聚合去重**(默认按 `IP:Port` 去重)。
- **富字段合并**:合并后的资产取各引擎字段的**并集(超集)**,而非公共交集;各引擎原始 JSON 无损保留在 `per_source`。
- **可配置 LLM 后端**:支持 OpenAI、OpenRouter 及任意 OpenAI 兼容端点(改 `base_url` 即可)。
- **引擎可启用/禁用**。
- **CLI**:支持 `-json` 结构化输出与彩色标准输出(`auto`/`always`/`never`)。
- **核心与展示解耦**:核心库(`pkg/search`)零 stdout / 零 `os.Exit`,便于将来做 Web。

目前内置编译器覆盖的引擎:**FOFA、Censys、Hunter、ZoomEye**(其余 uncover 引擎可按需扩展映射表)。

## 安装

```bash
go build -o uncover-turbo ./cmd/uncover-turbo
```

## 配置

### 引擎凭据

各测绘引擎的 API Key 沿用 **uncover 自身的配置方式**(环境变量 / provider-config),请按 [uncover 官方指导](https://github.com/projectdiscovery/uncover) 配置。

### 本工具配置文件(LLM 后端 / 引擎开关 / 输出)

默认路径 `~/.config/uncover-turbo/config.yaml`(可用 `-config` 指定):

```yaml
llm:
  base_url: "https://api.openai.com/v1"   # 换成 https://openrouter.ai/api/v1 即用 OpenRouter
  model:    "gpt-3.5-turbo"
  api_key:  ""                            # 建议留空,用环境变量覆盖
  temperature: 0.0
engines:
  enabled: [fofa, censys, hunter, zoomeye]
output:
  format: text                            # text | json
  color:  auto                            # auto | always | never
limit: 100
```

配置优先级(由低到高):**内置默认 < YAML 文件 < 环境变量 < 命令行参数**。

敏感/常改项可用环境变量覆盖:

| 环境变量 | 作用 |
|---|---|
| `UNCOVER_TURBO_LLM_API_KEY`  | LLM API Key |
| `UNCOVER_TURBO_LLM_BASE_URL` | LLM Base URL |
| `UNCOVER_TURBO_LLM_MODEL`    | LLM 模型名 |

## 使用

### 自然语言入口(`-q`,需配置 LLM)

```bash
export UNCOVER_TURBO_LLM_API_KEY=sk-xxx
./uncover-turbo -q '美国开放 3306 端口的主机' -e fofa,censys -v
```

`-v` 会在 stderr 打印翻译出的 IR 以及各引擎编译出的查询语法。

### 机器入口(`-ir`,绕过 LLM)

直接提供 IR(JSON),适合程序化调用;`-ir -` 可从 stdin 读取:

```bash
./uncover-turbo -ir '{"and":[
  {"match":{"field":"port","op":"eq","value":"3306"}},
  {"match":{"field":"country","op":"eq","value":"US"}}
]}' -json

echo '{"match":{"field":"title","op":"contains","value":"admin"}}' | ./uncover-turbo -ir -
```

### 常用参数

| 参数 | 说明 |
|---|---|
| `-q <text>`      | 自然语言查询(LLM 上层) |
| `-ir <json>`     | 直接提供 IR(`-` 表示从 stdin 读取) |
| `-e, -engines`   | 逗号分隔的引擎列表(覆盖配置) |
| `-config <path>` | 配置文件路径 |
| `-json`          | 输出 JSON |
| `-no-color`      | 关闭颜色 |
| `-l, -limit`     | 每引擎结果上限 |
| `-proxy`         | HTTP 代理 |
| `-v`             | 打印 IR 与各引擎查询到 stderr |

## 中间层 IR 参考

IR 是一棵布尔 AST,每个节点四选一:

```jsonc
{"and": [<expr>, ...]}                                  // 逻辑与(≥2 子节点)
{"or":  [<expr>, ...]}                                  // 逻辑或(≥2 子节点)
{"not": <expr>}                                          // 逻辑非
{"match": {"field": <字段>, "op": <eq|contains|ne>, "value": <字符串>}}
```

规范字段(引擎无关):`ip`、`port`、`domain`、`host`、`title`、`body`、`product`、`country`、`org`、`asn`、`protocol`、`status`、`cert.subject_cn`、`os`。

> 某引擎无法表达 IR 中的某字段/操作符时,该引擎会被**跳过并记录**(不发出错误查询),其余引擎照常返回。

## 开发

本仓库遵循 **SDD(规约驱动开发)红线**,流程与设计文档见 `specs/cross-engine-search/`(spec → plan → tasks)与根目录 `CLAUDE.md`。

```bash
go build ./...
go vet ./...
go test ./...
```

包结构:

```
cmd/uncover-turbo   薄 CLI 前端
pkg/queryir         中间层 IR(AST / 词汇 / 校验)
pkg/compiler        IR → 各引擎语法的确定性编译器(数据驱动 dialect)
pkg/llm             NL → IR 翻译器(OpenAI 兼容)
pkg/prompts         NL → IR 提示词
pkg/asset           资产模型 + 聚合去重
pkg/search          核心编排(SearchNL / SearchIR)
pkg/config          配置载入
pkg/render          JSON / 彩色文本渲染
```

## 说明

- 由 LLM 生成的查询可能存在语义偏差;`-v` 可查看翻译出的 IR 与各引擎查询便于核对。
- 各引擎字段映射为当前版本的最佳近似,以各引擎官方语法文档为准,可在 `pkg/compiler/engines.go` 增量校正。
- 本工具用于**授权范围内**的资产测绘与安全研究。
