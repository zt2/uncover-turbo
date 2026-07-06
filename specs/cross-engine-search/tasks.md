# Tasks · 自然语言跨多测绘引擎查询与聚合去重

> SDD 阶段三产出物。把已确认的 `plan.md` 拆为有序、可独立验证的小任务。
> **状态:已确认(含里程碑排序调整)。** 进入阶段四（Implement)。
> 每条完成后运行 `go build ./... && go vet ./...`(带测试的加 `go test ./...`)。

## 里程碑排序(已确认)

> 决策:**先做底层聚合 + 中间层 IR + IR→引擎语法翻译;上层 LLM 自然语言翻译放到后面。**

- **里程碑 A(先做)**:T0、T1、T2、T3–T6、T7、T10、**T11(仅 `SearchIR`)**、T12、**T13(仅 `-ir` 机器入口)**、T15A 终验。
  产物:一个**不依赖 LLM**、可通过 `-ir` 直接注入 IR 的跨引擎查询/聚合/渲染工具。
- **里程碑 B(后做)**:T8、T9、T11 的 `SearchNL`、T13 的 `-q`(自然语言入口)、T14 README、T15B 终验。
  产物:补上「自然语言 → IR」的 LLM 上层。

## 依赖关系概览

```
T0 脚手架
 └─ T1 queryir ─┬─ T2 compiler 接口 ─┬─ T3 fofa
                │                     ├─ T4 censys
                │                     ├─ T5 hunter
                │                     └─ T6 zoomeye
                ├─ T8 prompts ── T9 llm
                └───────────────┐
T7 asset(独立) ────────────────┤
T10 config(独立) ──────────────┼─ T11 search ─ T12 render ─ T13 cli ─ T14 清理/README ─ T15 终验
```

---

## T0 · 脚手架与依赖
- **动作**:`go get gopkg.in/yaml.v3`;建立 §10 包目录骨架(空文件/占位)。不改旧代码。
- **验收**:`go build ./...` 通过;`go.mod` 含 yaml.v3。

## T1 · pkg/queryir(中间层 IR)
- **动作**:定义 `Expr`(And/Or/Not/Match 四选一)、`Match`、`Field` 规范词汇(§2.2 的 14 字段)、`Op`(eq/contains/ne);实现 `Validate()`、`String()`;JSON tag 完备。
- **验收**:单测通过——JSON 往返一致;非法结构(多字段同设/空 value/未知 field)被 `Validate` 拒;`String()` 输出符合预期。**依赖 T0**。

## T2 · pkg/compiler(编译器接口 + 注册表)
- **动作**:`Compiler` 接口(`Engine()`, `Compile(Expr)(string,error)`);`ErrUnsupported`;`registry`(按引擎名取编译器);数据驱动 `dialect`(通用布尔遍历 + 取反模型 + 叶子格式化)。
- **验收**:接口与注册表编译通过;单测验证注册/取用与不支持返回。**依赖 T1**。
- **实现修订**:四引擎不再各自子包,而是 `pkg/compiler/engines.go` 内四个 `dialect` 值(字段映射 + 布尔语法),由 `dialect.go` 的通用编译逻辑驱动。

## T3–T6 · 四引擎 dialect(fofa / censys / hunter / zoomeye)
- **动作**:在 `engines.go` 为四家各定义 `dialect`——字段映射表 + 布尔/取反语法:
  - fofa/hunter:`field="value"`、`&&`/`||`、字段级 `!=`、分组 `()`;
  - censys:`field: value`、`and`/`or`、一元 `not (...)`、分组 `()`;
  - zoomeye:`key:"value"`、空格隐式 AND、`-` 词级取反(无 OR / 无分组取反 → `ErrUnsupported`)。
- **验收**:表驱动单测断言**精确**查询串 + 不支持字段/op 返回 `ErrUnsupported` + 值转义。**依赖 T2**。

## T7 · pkg/asset(模型 + 聚合去重)
- **动作**:`Asset` 结构(§5);`Aggregator`——`Add(sources.Result)` 按 `IP:Port`(空 IP 退化 `host:Port`)合并;`Assets()` 返回稳定排序结果。合并规则:Hosts/URLs/Sources 并集去重排序;`Fields` 解析各 Raw 扁平并入(首个非空优先);`PerSource[source]=Raw` 无损。
- **验收**:单测——同 IP:Port 多引擎结果合并为一条;Fields 取并集且首个非空优先;Sources/Hosts 并集;PerSource 无损;不同资产不误并。覆盖 **AC2/AC3**。**独立(仅依赖 uncover sources 类型)**。

## T8 · pkg/prompts(NL→IR 单一提示词)
- **动作**:一个 system prompt:说明 IR 的 JSON schema、规范字段词汇、示例、只准输出 JSON。
- **验收**:prompt 常量存在并被 llm 引用;go vet 通过。**依赖 T1(词汇引用)**。

## T9 · pkg/llm(翻译器 NL→IR)
- **动作**:`Translator` 接口;`openAITranslator`(go-openai + 自定义 BaseURL/模型/温度);回复 `json.Unmarshal`→`queryir.Validate`,失败把错误回喂重试一次。把「发起对话」抽象为内部小接口以便注入 fake。
- **验收**:单测(注入 fake chat)——合法 JSON→得到有效 IR;非法→重试一次→仍非法则明确报错。不联网。**依赖 T1、T8**。

## T10 · pkg/config
- **动作**:`Config` 结构 + `Load(path)`:读 YAML(缺文件用默认)、环境变量覆盖(`UNCOVER_TURBO_LLM_*`)、暴露给 CLI 覆盖的字段;优先级 CLI>env>文件>默认(CLI 覆盖在 main 里做,config 负责 env>文件>默认)。
- **验收**:单测——YAML 载入正确;env 覆盖文件;缺文件回落默认;引擎 enabled 列表解析正确。**独立**。

## T11 · pkg/search(核心编排)
- **动作**:`Service`(translator + compilers + config);`SearchIR`(校验 IR → 各启用引擎 `Compile`(不支持则记 EngineErrors 跳过)→ 每引擎独立 uncover Execute → 汇入 `asset.Aggregator`);`SearchNL`(translate→SearchIR);`Result{Assets,IR,Queries,EngineErrors}`。把「引擎查询」抽象为 `engineQuerier` 接口,默认实现用 uncover,测试用 fake。
- **验收**:单测(fake translator + fake querier)——编译分发正确;不支持引擎被跳过并记错;多引擎结果聚合去重;`SearchIR` 可脱离 LLM 独立工作(**AC9**);核心无 stdout/os.Exit(**AC7**)。覆盖 AC1/AC5。**依赖 T2–T7、T9、T10**。

## T12 · pkg/render
- **动作**:`Renderer` 接口;`jsonRenderer`(每资产一对象 + 可选 meta:IR/queries);`textRenderer`(aurora 上色,`auto` 仅 TTY / `always` / `never`)。
- **验收**:单测——给定固定 Result,JSON 结构正确;`never` 下文本无 ANSI 码。覆盖 **AC4**。**依赖 T11(Result 类型)**。

## T13 · cmd/uncover-turbo/main.go(薄 CLI)
- **动作**:重写。flags `-q`/`-ir`(二选一)/`-e`/`-config`/`-json`/`-no-color`/`-l`/`-proxy`/`-v`/`-delay`;`config.Load`+flag 覆盖;装配 Translator/Compilers/Service;`-ir`→SearchIR 否则 SearchNL;选 Renderer 输出;`-v` 打印 IR 与各引擎 query;按致命错误定退出码。
- **验收**:`go build ./...` 出二进制;`-h` 正常;`-v -ir '<json>'` 走通编排(无 Key 时引擎降级、聚合空结果、不崩溃)。覆盖 AC1/AC9 可观测。**依赖 T11、T12**。

## T14 · 清理与文档
- **动作**:删除旧 `pkg/bot`;更新 `README.md`(中文)——新三层架构、配置说明(YAML + env + 引擎凭据走 uncover)、IR 示例与 `-ir` 机器调用、多引擎/聚合/富字段/JSON/颜色用法、示例更新。
- **验收**:无对已删包的悬挂引用;`go build ./...` 通过;README 覆盖新用法。**依赖 T13**。

## T15 · 终验
- **动作**:全量 `go build ./... && go vet ./... && go test ./...`;必要时 `go mod tidy`;跑 `/verify` 或等价冒烟(构建二进制 + `-h` + 一次 `-ir` 干跑)。
- **验收**:三门禁全绿;二进制可运行;spec 全部 AC 有对应实现/测试或可观测证据。**依赖 T14**。

---

### 提交策略
按任务粒度提交(每 1~2 个相关任务一次 commit,信息可追溯到任务号与规约),便于独立回滚。全部完成后推送并开 PR 至 `develop`。

### 后续
确认本 Tasks 后进入 **阶段四 · Implement**,按 T0→T15 顺序实现。
