package prompts

import (
	"strings"
	"testing"

	"github.com/zt2/uncover-turbo/pkg/queryir"
)

func TestSystemMentionsEveryField(t *testing.T) {
	p := System()
	for _, f := range queryir.Fields() {
		if !strings.Contains(p, string(f)) {
			t.Errorf("system prompt missing field %q", f)
		}
	}
	// Sanity: mentions the four node kinds and JSON-only instruction.
	for _, kw := range []string{"and", "or", "not", "match", "JSON"} {
		if !strings.Contains(p, kw) {
			t.Errorf("system prompt missing keyword %q", kw)
		}
	}
}
