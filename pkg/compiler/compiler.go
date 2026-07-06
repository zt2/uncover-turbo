// Package compiler is the lower layer of the pipeline: it deterministically
// compiles an engine-agnostic queryir.Expr into a specific recon engine's
// proprietary query syntax. No LLM is involved — the same IR always compiles to
// the same query string for a given engine.
//
// Each engine is expressed as a data-driven dialect (field mapping + boolean
// syntax), so adding or correcting an engine is a table edit rather than new
// control flow.
package compiler

import (
	"errors"
	"sort"

	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// ErrUnsupported is returned (wrapped) when an engine cannot express a field or
// operator present in the IR. Callers should treat this as "skip this engine for
// this query" rather than a fatal error.
var ErrUnsupported = errors.New("compiler: field or operator not supported by engine")

// Compiler turns an IR expression into one engine's query syntax.
type Compiler interface {
	Engine() string
	Compile(expr queryir.Expr) (string, error)
}

// registry holds the built-in engine compilers, keyed by engine name.
var registry = map[string]Compiler{}

// register adds c to the registry. Called from init in engines.go.
func register(c Compiler) { registry[c.Engine()] = c }

// Get returns the compiler for the named engine.
func Get(engine string) (Compiler, bool) {
	c, ok := registry[engine]
	return c, ok
}

// Engines returns the names of all registered engines, sorted.
func Engines() []string {
	out := make([]string, 0, len(registry))
	for name := range registry {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
