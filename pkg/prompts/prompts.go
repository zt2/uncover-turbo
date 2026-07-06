// Package prompts builds the single system prompt used by the LLM upper layer to
// translate natural language into the query IR. The field list is generated from
// the live queryir vocabulary so the prompt cannot drift from the code.
package prompts

import (
	"fmt"
	"strings"

	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// fieldDoc gives a short human description for each canonical field, used to
// help the model pick the right one.
var fieldDoc = map[queryir.Field]string{
	queryir.FieldIP:       "IP 地址",
	queryir.FieldPort:     "端口",
	queryir.FieldDomain:   "注册域名",
	queryir.FieldHost:     "主机名 / hostname",
	queryir.FieldTitle:    "HTTP 页面标题",
	queryir.FieldBody:     "响应体 / 正文关键字",
	queryir.FieldProduct:  "组件 / 产品名称",
	queryir.FieldCountry:  "国家码,如 US、CN",
	queryir.FieldOrg:      "组织名",
	queryir.FieldASN:      "自治系统号 ASN",
	queryir.FieldProtocol: "服务 / 协议,如 http、ssh",
	queryir.FieldStatus:   "HTTP 状态码",
	queryir.FieldCertCN:   "TLS 证书 subject CN",
	queryir.FieldOS:       "操作系统",
}

// System returns the NL->IR system prompt.
func System() string {
	var fields strings.Builder
	for _, f := range queryir.Fields() {
		fmt.Fprintf(&fields, "  - %q: %s\n", string(f), fieldDoc[f])
	}

	return `你是一个把「自然语言测绘需求」翻译成「通用查询 IR」的转换器。

只输出一个 JSON 对象,不要输出任何解释、注释或 Markdown 代码块围栏。

IR 是一棵布尔表达式树,每个节点必须且只能是以下四种之一:
  - {"and": [<expr>, <expr>, ...]}   // 逻辑与,至少 2 个子节点
  - {"or":  [<expr>, <expr>, ...]}   // 逻辑或,至少 2 个子节点
  - {"not": <expr>}                   // 逻辑非
  - {"match": {"field": <字段>, "op": <操作符>, "value": <字符串>}}  // 叶子谓词

叶子谓词的 op 只能是:
  - "eq":       等于 / 精确匹配
  - "contains": 包含 / 关键字匹配
  - "ne":       不等于

field 只能取以下规范字段之一(不要臆造其它字段名):
` + fields.String() + `
value 一律为字符串(端口、状态码等数字也写成字符串,如 "3306")。

示例:
自然语言:美国开放 3306 端口的主机
输出:{"and":[{"match":{"field":"port","op":"eq","value":"3306"}},{"match":{"field":"country","op":"eq","value":"US"}}]}

自然语言:标题包含 admin 或正文包含 login 的网站
输出:{"or":[{"match":{"field":"title","op":"contains","value":"admin"}},{"match":{"field":"body","op":"contains","value":"login"}}]}

现在请把用户输入翻译成 IR,只输出 JSON。`
}
