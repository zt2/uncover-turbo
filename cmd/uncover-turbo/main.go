// Command uncover-turbo runs cross-engine recon queries and aggregates the
// results into deduplicated assets.
//
// It is a thin front-end over pkg/search: it parses flags, loads config, invokes
// the core, and renders the result. Two entry points are supported:
//
//	-ir '<json>'   machine entry: supply an engine-agnostic IR directly
//	-q  '<text>'   natural-language entry (LLM upper layer; milestone B)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	_ "github.com/projectdiscovery/fdmax/autofdmax"
	"github.com/zt2/uncover-turbo/pkg/config"
	"github.com/zt2/uncover-turbo/pkg/queryir"
	"github.com/zt2/uncover-turbo/pkg/render"
	"github.com/zt2/uncover-turbo/pkg/search"
)

type options struct {
	ir       string
	nl       string
	engines  string
	confPath string
	jsonOut  bool
	noColor  bool
	limit    int
	proxy    string
	verbose  bool
}

func main() {
	opts := parseFlags()

	cfg, err := config.Load(configPath(opts.confPath))
	if err != nil {
		fatal("load config: %v", err)
	}
	applyOverrides(cfg, opts)

	svc := search.New(cfg)
	ctx := context.Background()

	var res *search.Result
	switch {
	case opts.ir != "":
		ir, perr := parseIR(opts.ir)
		if perr != nil {
			fatal("parse -ir: %v", perr)
		}
		res, err = svc.SearchIR(ctx, ir)
	case opts.nl != "":
		res, err = svc.SearchNL(ctx, opts.nl)
	default:
		fatal("provide a query: -ir '<json>' (machine) or -q '<text>' (natural language)")
	}
	if err != nil {
		fatal("%v", err)
	}

	if opts.verbose {
		printMeta(os.Stderr, res)
	}
	printEngineWarnings(os.Stderr, res)

	if err := renderer(cfg, opts).Render(os.Stdout, res); err != nil {
		fatal("render: %v", err)
	}
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.ir, "ir", "", "engine-agnostic IR query as JSON (use '-' to read from stdin)")
	flag.StringVar(&o.nl, "q", "", "natural-language query (LLM upper layer)")
	flag.StringVar(&o.engines, "e", "", "comma-separated engines to query (overrides config)")
	flag.StringVar(&o.engines, "engines", "", "comma-separated engines to query (overrides config)")
	flag.StringVar(&o.confPath, "config", "", "config file path (default ~/.config/uncover-turbo/config.yaml)")
	flag.BoolVar(&o.jsonOut, "json", false, "output JSON instead of text")
	flag.BoolVar(&o.noColor, "no-color", false, "disable colored output")
	flag.IntVar(&o.limit, "l", -1, "max results per engine (overrides config)")
	flag.IntVar(&o.limit, "limit", -1, "max results per engine (overrides config)")
	flag.StringVar(&o.proxy, "proxy", "", "HTTP proxy for engine queries")
	flag.BoolVar(&o.verbose, "v", false, "verbose: print IR and per-engine queries to stderr")
	flag.Parse()
	return o
}

// configPath returns the explicit path if set, otherwise the default location.
func configPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return config.DefaultPath()
}

// applyOverrides layers CLI flags on top of the loaded config (highest priority).
func applyOverrides(cfg *config.Config, o options) {
	if o.engines != "" {
		cfg.Engines.Enabled = splitCSV(o.engines)
	}
	if o.limit >= 0 {
		cfg.Limit = o.limit
	}
	if o.proxy != "" {
		cfg.Proxy = o.proxy
	}
	if o.jsonOut {
		cfg.Output.Format = "json"
	}
	if o.noColor {
		cfg.Output.Color = "never"
	}
}

func renderer(cfg *config.Config, o options) render.Renderer {
	if cfg.Output.Format == "json" {
		// Include meta (IR, per-engine queries, errors) when verbose.
		return render.NewJSON(o.verbose, true)
	}
	return render.NewText(render.ResolveColor(cfg.Output.Color, os.Stdout))
}

// parseIR reads the IR JSON (from the string or stdin when "-") and validates it.
func parseIR(src string) (queryir.Expr, error) {
	var data []byte
	if src == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return queryir.Expr{}, fmt.Errorf("read stdin: %w", err)
		}
		data = b
	} else {
		data = []byte(src)
	}
	var expr queryir.Expr
	if err := json.Unmarshal(data, &expr); err != nil {
		return queryir.Expr{}, fmt.Errorf("invalid IR JSON: %w", err)
	}
	if err := expr.Validate(); err != nil {
		return queryir.Expr{}, err
	}
	return expr, nil
}

func printMeta(w io.Writer, res *search.Result) {
	fmt.Fprintf(w, "IR: %s\n", res.IR.String())
	for engine, q := range res.Queries {
		fmt.Fprintf(w, "  [%s] %s\n", engine, q)
	}
}

func printEngineWarnings(w io.Writer, res *search.Result) {
	for engine, err := range res.EngineErrors {
		fmt.Fprintf(w, "[warn] %s skipped: %v\n", engine, err)
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "uncover-turbo: "+format+"\n", args...)
	os.Exit(1)
}
