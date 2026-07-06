package render

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/zt2/uncover-turbo/pkg/search"
)

// TextRenderer emits one human-readable line per asset:
//
//	<ip>:<port>  host1,host2  [source1,source2]
//
// Coloring is controlled at construction; when disabled no ANSI codes are
// emitted, so piped/redirected output stays clean.
type TextRenderer struct {
	au aurora.Aurora
}

// NewText returns a text renderer. Pass color=false to disable ANSI coloring.
func NewText(color bool) *TextRenderer {
	return &TextRenderer{au: aurora.NewAurora(color)}
}

func (r *TextRenderer) Render(w io.Writer, res *search.Result) error {
	for i := range res.Assets {
		a := res.Assets[i]

		target := a.IP
		if target == "" && len(a.Hosts) > 0 {
			target = a.Hosts[0]
		}
		addr := target + ":" + strconv.Itoa(a.Port)

		var b strings.Builder
		b.WriteString(fmt.Sprintf("%v", r.au.Bold(r.au.Cyan(addr))))
		if len(a.Hosts) > 0 {
			b.WriteString("  ")
			b.WriteString(strings.Join(a.Hosts, ","))
		}
		if len(a.Sources) > 0 {
			b.WriteString("  ")
			b.WriteString(fmt.Sprintf("%v", r.au.Green("["+strings.Join(a.Sources, ",")+"]")))
		}
		if _, err := fmt.Fprintln(w, b.String()); err != nil {
			return err
		}
	}
	return nil
}
