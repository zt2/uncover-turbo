package render

import (
	"encoding/json"
	"io"

	"github.com/zt2/uncover-turbo/pkg/asset"
	"github.com/zt2/uncover-turbo/pkg/queryir"
	"github.com/zt2/uncover-turbo/pkg/search"
)

// JSONRenderer emits the result as a single structured JSON object. When
// withMeta is set, a "meta" section (IR, per-engine queries, engine errors) is
// included alongside the assets.
type JSONRenderer struct {
	WithMeta bool
	Indent   bool
}

// NewJSON returns a JSON renderer.
func NewJSON(withMeta, indent bool) *JSONRenderer {
	return &JSONRenderer{WithMeta: withMeta, Indent: indent}
}

type jsonMeta struct {
	IR           queryir.Expr      `json:"ir"`
	Queries      map[string]string `json:"queries,omitempty"`
	EngineErrors map[string]string `json:"engine_errors,omitempty"`
}

type jsonOutput struct {
	Meta   *jsonMeta     `json:"meta,omitempty"`
	Assets []asset.Asset `json:"assets"`
}

func (r *JSONRenderer) Render(w io.Writer, res *search.Result) error {
	out := jsonOutput{Assets: res.Assets}
	if out.Assets == nil {
		out.Assets = []asset.Asset{}
	}
	if r.WithMeta {
		out.Meta = &jsonMeta{
			IR:           res.IR,
			Queries:      res.Queries,
			EngineErrors: stringifyErrors(res.EngineErrors),
		}
	}

	enc := json.NewEncoder(w)
	if r.Indent {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// stringifyErrors converts an engine->error map into an engine->string map for
// JSON, returning nil when empty so the field is omitted.
func stringifyErrors(errs map[string]error) map[string]string {
	if len(errs) == 0 {
		return nil
	}
	out := make(map[string]string, len(errs))
	for engine, err := range errs {
		out[engine] = err.Error()
	}
	return out
}
