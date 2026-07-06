package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/zt2/uncover-turbo/pkg/asset"
	"github.com/zt2/uncover-turbo/pkg/queryir"
	"github.com/zt2/uncover-turbo/pkg/search"
)

func sampleResult() *search.Result {
	return &search.Result{
		IR: queryir.Expr{Match: &queryir.Match{Field: queryir.FieldPort, Op: queryir.OpEq, Value: "3306"}},
		Assets: []asset.Asset{{
			IP:      "1.2.3.4",
			Port:    3306,
			Hosts:   []string{"a.com", "b.com"},
			Sources: []string{"censys", "fofa"},
			Fields:  map[string]any{"product": "mysql"},
		}},
		Queries:      map[string]string{"fofa": `port="3306"`},
		EngineErrors: map[string]error{"zoomeye": errors.New("unsupported")},
	}
}

func TestJSONStructure(t *testing.T) {
	var buf bytes.Buffer
	if err := NewJSON(true, false).Render(&buf, sampleResult()); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Meta struct {
			Queries      map[string]string `json:"queries"`
			EngineErrors map[string]string `json:"engine_errors"`
			IR           json.RawMessage   `json:"ir"`
		} `json:"meta"`
		Assets []struct {
			IP      string   `json:"ip"`
			Port    int      `json:"port"`
			Sources []string `json:"sources"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(out.Assets) != 1 || out.Assets[0].IP != "1.2.3.4" || out.Assets[0].Port != 3306 {
		t.Errorf("asset not rendered: %+v", out.Assets)
	}
	if out.Meta.Queries["fofa"] != `port="3306"` {
		t.Errorf("meta queries wrong: %v", out.Meta.Queries)
	}
	if out.Meta.EngineErrors["zoomeye"] != "unsupported" {
		t.Errorf("meta engine_errors wrong: %v", out.Meta.EngineErrors)
	}
}

func TestJSONWithoutMetaOmitsMeta(t *testing.T) {
	var buf bytes.Buffer
	if err := NewJSON(false, false).Render(&buf, sampleResult()); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), `"meta"`) {
		t.Errorf("meta should be omitted:\n%s", buf.String())
	}
}

func TestJSONEmptyAssetsIsArray(t *testing.T) {
	var buf bytes.Buffer
	res := &search.Result{}
	if err := NewJSON(false, false).Render(&buf, res); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"assets":[]`) {
		t.Errorf("empty assets should serialize as []:\n%s", buf.String())
	}
}

func TestTextNoColorHasNoANSI(t *testing.T) {
	var buf bytes.Buffer
	if err := NewText(false).Render(&buf, sampleResult()); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if strings.Contains(s, "\x1b[") {
		t.Errorf("no-color output must not contain ANSI escapes: %q", s)
	}
	if !strings.Contains(s, "1.2.3.4:3306") {
		t.Errorf("expected address in output: %q", s)
	}
	if !strings.Contains(s, "[censys,fofa]") {
		t.Errorf("expected sources in output: %q", s)
	}
}

func TestTextColorHasANSI(t *testing.T) {
	var buf bytes.Buffer
	if err := NewText(true).Render(&buf, sampleResult()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("color output should contain ANSI escapes: %q", buf.String())
	}
}

func TestResolveColor(t *testing.T) {
	// A bytes.Buffer is not a terminal, so auto -> false.
	var buf bytes.Buffer
	if ResolveColor("auto", &buf) {
		t.Error("auto on non-terminal should be false")
	}
	if !ResolveColor("always", &buf) {
		t.Error("always should be true")
	}
	if ResolveColor("never", &buf) {
		t.Error("never should be false")
	}
}
