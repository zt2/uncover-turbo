package main

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/zt2/uncover-turbo/pkg/config"
	"github.com/zt2/uncover-turbo/pkg/queryir"
	"github.com/zt2/uncover-turbo/pkg/render"
	"github.com/zt2/uncover-turbo/pkg/search"
)

func TestSplitCSV(t *testing.T) {
	cases := map[string][]string{
		"fofa,censys":     {"fofa", "censys"},
		" fofa , censys ": {"fofa", "censys"}, // trims whitespace
		"fofa,,censys":    {"fofa", "censys"}, // drops empties
		"fofa":            {"fofa"},
		"":                {},
		"  ":              {},
		",":               {},
	}
	for in, want := range cases {
		got := splitCSV(in)
		if len(want) == 0 && len(got) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitCSV(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestConfigPath(t *testing.T) {
	if got := configPath("/explicit/path.yaml"); got != "/explicit/path.yaml" {
		t.Errorf("explicit path not honored: %q", got)
	}
	// Empty falls back to the package default (may be "" if no home, but must
	// match config.DefaultPath()).
	if got := configPath(""); got != config.DefaultPath() {
		t.Errorf("empty should use DefaultPath(): %q vs %q", got, config.DefaultPath())
	}
}

func TestApplyOverridesPrecedence(t *testing.T) {
	cfg := config.Default()
	base := config.Default()

	// No overrides: config unchanged.
	applyOverrides(cfg, options{limit: -1})
	if !reflect.DeepEqual(cfg.Engines.Enabled, base.Engines.Enabled) || cfg.Limit != base.Limit {
		t.Fatalf("no-op overrides changed config: %+v", cfg)
	}

	// Flags win over the loaded config.
	applyOverrides(cfg, options{
		engines: "fofa,zoomeye",
		limit:   7,
		proxy:   "http://p:8080",
		jsonOut: true,
		noColor: true,
	})
	if !reflect.DeepEqual(cfg.Engines.Enabled, []string{"fofa", "zoomeye"}) {
		t.Errorf("engines override failed: %v", cfg.Engines.Enabled)
	}
	if cfg.Limit != 7 || cfg.Proxy != "http://p:8080" {
		t.Errorf("limit/proxy override failed: %+v", cfg)
	}
	if cfg.Output.Format != "json" || cfg.Output.Color != "never" {
		t.Errorf("json/no-color override failed: %+v", cfg.Output)
	}
}

func TestApplyOverridesLimitSentinel(t *testing.T) {
	cfg := config.Default()
	cfg.Limit = 42
	applyOverrides(cfg, options{limit: -1}) // -1 means "unset": keep config value
	if cfg.Limit != 42 {
		t.Errorf("limit -1 sentinel should not override, got %d", cfg.Limit)
	}
	applyOverrides(cfg, options{limit: 0}) // explicit 0 is a real value
	if cfg.Limit != 0 {
		t.Errorf("limit 0 should override, got %d", cfg.Limit)
	}
}

func TestRendererSelection(t *testing.T) {
	jsonCfg := config.Default()
	jsonCfg.Output.Format = "json"
	if _, ok := renderer(jsonCfg, options{}).(*render.JSONRenderer); !ok {
		t.Error("json format should yield *render.JSONRenderer")
	}
	textCfg := config.Default()
	textCfg.Output.Format = "text"
	if _, ok := renderer(textCfg, options{}).(*render.TextRenderer); !ok {
		t.Error("text format should yield *render.TextRenderer")
	}
}

func TestParseIRString(t *testing.T) {
	expr, err := parseIR(`{"match":{"field":"port","op":"eq","value":"3306"}}`)
	if err != nil {
		t.Fatalf("valid IR string: %v", err)
	}
	if expr.Match == nil || expr.Match.Value != "3306" {
		t.Errorf("wrong parse: %+v", expr)
	}
}

func TestParseIRErrors(t *testing.T) {
	if _, err := parseIR(`{not valid json`); err == nil {
		t.Error("expected JSON error")
	}
	// Valid JSON but invalid IR (unknown field) -> validation error.
	if _, err := parseIR(`{"match":{"field":"bogus","op":"eq","value":"x"}}`); err == nil {
		t.Error("expected IR validation error")
	}
}

func TestPrintMeta(t *testing.T) {
	var buf bytes.Buffer
	res := &search.Result{
		IR:      queryir.Expr{Match: &queryir.Match{Field: queryir.FieldPort, Op: queryir.OpEq, Value: "3306"}},
		Queries: map[string]string{"fofa": `port="3306"`},
	}
	printMeta(&buf, res)
	out := buf.String()
	if !strings.Contains(out, `IR: port eq "3306"`) {
		t.Errorf("meta missing IR line: %q", out)
	}
	if !strings.Contains(out, `[fofa] port="3306"`) {
		t.Errorf("meta missing per-engine query: %q", out)
	}
}

func TestPrintEngineWarnings(t *testing.T) {
	var buf bytes.Buffer
	res := &search.Result{EngineErrors: map[string]error{"zoomeye": errors.New("unsupported OR")}}
	printEngineWarnings(&buf, res)
	if !strings.Contains(buf.String(), "[warn] zoomeye skipped: unsupported OR") {
		t.Errorf("missing warning line: %q", buf.String())
	}

	// No errors -> no output.
	buf.Reset()
	printEngineWarnings(&buf, &search.Result{})
	if buf.Len() != 0 {
		t.Errorf("expected no output for no errors, got %q", buf.String())
	}
}

func TestParseIRStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = orig }()

	go func() {
		_, _ = w.Write([]byte(`{"match":{"field":"ip","op":"eq","value":"1.1.1.1"}}`))
		_ = w.Close()
	}()

	expr, err := parseIR("-")
	if err != nil {
		t.Fatalf("stdin IR: %v", err)
	}
	if expr.Match == nil || expr.Match.Value != "1.1.1.1" {
		t.Errorf("wrong stdin parse: %+v", expr)
	}
}
