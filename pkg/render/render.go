// Package render turns a search.Result into output: structured JSON or a
// human-readable, optionally colored, text form. It is the presentation layer;
// the core (pkg/search) never writes output itself.
package render

import (
	"io"
	"os"

	"github.com/zt2/uncover-turbo/pkg/search"
)

// Renderer writes a search result to w.
type Renderer interface {
	Render(w io.Writer, res *search.Result) error
}

// ResolveColor decides whether color should be used for the given color mode
// ("auto" | "always" | "never") and output writer. In auto mode color is on
// only when w is a terminal.
func ResolveColor(mode string, w io.Writer) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	default: // auto
		return isTerminal(w)
	}
}

// isTerminal reports whether w is a character device (a reasonable proxy for a
// TTY) without pulling in an extra dependency.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
