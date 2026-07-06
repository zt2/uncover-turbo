package llm

import (
	"context"
	"errors"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// scriptedCompleter returns queued responses (and/or an error) in order,
// recording how many times it was called.
type scriptedCompleter struct {
	responses []string
	err       error
	calls     int
}

func (s *scriptedCompleter) complete(ctx context.Context, _ []openai.ChatCompletionMessage) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	i := s.calls
	s.calls++
	if i < len(s.responses) {
		return s.responses[i], nil
	}
	return s.responses[len(s.responses)-1], nil
}

func TestTranslateValidFirstTry(t *testing.T) {
	c := &scriptedCompleter{responses: []string{
		`{"and":[{"match":{"field":"port","op":"eq","value":"3306"}},{"match":{"field":"country","op":"eq","value":"US"}}]}`,
	}}
	tr := newWithCompleter(c)
	expr, err := tr.Translate(context.Background(), "美国 3306")
	if err != nil {
		t.Fatal(err)
	}
	if err := expr.Validate(); err != nil {
		t.Fatalf("returned IR invalid: %v", err)
	}
	if c.calls != 1 {
		t.Errorf("expected 1 call, got %d", c.calls)
	}
	want := `(port eq "3306" AND country eq "US")`
	if expr.String() != want {
		t.Errorf("got %q want %q", expr.String(), want)
	}
}

func TestTranslateStripsFences(t *testing.T) {
	c := &scriptedCompleter{responses: []string{
		"```json\n{\"match\":{\"field\":\"ip\",\"op\":\"eq\",\"value\":\"1.1.1.1\"}}\n```",
	}}
	tr := newWithCompleter(c)
	expr, err := tr.Translate(context.Background(), "ip 1.1.1.1")
	if err != nil {
		t.Fatalf("fenced JSON should parse: %v", err)
	}
	if expr.Match == nil || expr.Match.Value != "1.1.1.1" {
		t.Errorf("wrong parse: %+v", expr)
	}
}

func TestTranslateRetriesOnceThenSucceeds(t *testing.T) {
	c := &scriptedCompleter{responses: []string{
		`not json at all`,
		`{"match":{"field":"port","op":"eq","value":"22"}}`,
	}}
	tr := newWithCompleter(c)
	expr, err := tr.Translate(context.Background(), "ssh")
	if err != nil {
		t.Fatalf("should succeed on retry: %v", err)
	}
	if c.calls != 2 {
		t.Errorf("expected 2 calls (one retry), got %d", c.calls)
	}
	if expr.Match == nil || expr.Match.Value != "22" {
		t.Errorf("wrong parse after retry: %+v", expr)
	}
}

func TestTranslateFailsAfterRetry(t *testing.T) {
	c := &scriptedCompleter{responses: []string{`garbage`, `{"bogus":true}`}}
	tr := newWithCompleter(c)
	if _, err := tr.Translate(context.Background(), "x"); err == nil {
		t.Error("expected error after exhausting retries")
	}
	if c.calls != 2 {
		t.Errorf("expected exactly 2 attempts, got %d", c.calls)
	}
}

func TestTranslateInvalidFieldRejected(t *testing.T) {
	// Valid JSON shape but unknown field -> validation fails both attempts.
	c := &scriptedCompleter{responses: []string{
		`{"match":{"field":"nope","op":"eq","value":"x"}}`,
	}}
	tr := newWithCompleter(c)
	if _, err := tr.Translate(context.Background(), "x"); err == nil {
		t.Error("expected validation error for unknown field")
	}
}

func TestTranslateChatErrorPropagates(t *testing.T) {
	c := &scriptedCompleter{err: errors.New("network down")}
	tr := newWithCompleter(c)
	if _, err := tr.Translate(context.Background(), "x"); err == nil {
		t.Error("expected chat error to propagate")
	}
}

// Ensure Translator satisfies the shape expected by pkg/search (Translate).
var _ interface {
	Translate(context.Context, string) (queryir.Expr, error)
} = (*Translator)(nil)
