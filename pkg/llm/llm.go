// Package llm is the upper layer of the pipeline: it translates natural language
// into the query IR using an OpenAI-compatible chat model. A custom base URL
// makes it work with OpenAI, OpenRouter or any compatible endpoint.
//
// The actual chat call is behind the completer interface so the translation,
// validation and retry logic can be unit-tested without network access.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/zt2/uncover-turbo/pkg/config"
	"github.com/zt2/uncover-turbo/pkg/prompts"
	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// completer performs a single chat completion, returning the assistant message
// content.
type completer interface {
	complete(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error)
}

// Translator turns natural language into an IR expression.
type Translator struct {
	c completer
}

// New builds a Translator backed by an OpenAI-compatible endpoint described by
// cfg. cfg.BaseURL selects the provider (OpenAI, OpenRouter, ...).
func New(cfg config.LLM) *Translator {
	oc := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		oc.BaseURL = cfg.BaseURL
	}
	model := cfg.Model
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	return &Translator{
		c: &openAICompleter{
			client: openai.NewClientWithConfig(oc),
			model:  model,
			temp:   cfg.Temperature,
		},
	}
}

// newWithCompleter is used by tests to inject a fake completer.
func newWithCompleter(c completer) *Translator {
	return &Translator{c: c}
}

// Translate converts nl into a validated IR expression. If the model's first
// response cannot be parsed/validated, it retries once, feeding the error back
// so the model can correct itself.
func (t *Translator) Translate(ctx context.Context, nl string) (queryir.Expr, error) {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: prompts.System()},
		{Role: openai.ChatMessageRoleUser, Content: nl},
	}

	const maxAttempts = 2
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		content, err := t.c.complete(ctx, messages)
		if err != nil {
			return queryir.Expr{}, fmt.Errorf("llm: chat completion: %w", err)
		}
		expr, perr := parseIR(content)
		if perr == nil {
			return expr, nil
		}
		lastErr = perr
		// Feed the bad output and the error back for one corrective retry.
		messages = append(messages,
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: content},
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("上次输出无法解析为合法 IR:%s。请只输出合法的 JSON IR,不要任何多余内容。", perr.Error()),
			},
		)
	}
	return queryir.Expr{}, fmt.Errorf("llm: could not obtain a valid IR after %d attempts: %w", maxAttempts, lastErr)
}

// parseIR strips optional Markdown fences and parses+validates the IR JSON.
func parseIR(content string) (queryir.Expr, error) {
	s := stripFences(strings.TrimSpace(content))
	var expr queryir.Expr
	if err := json.Unmarshal([]byte(s), &expr); err != nil {
		return queryir.Expr{}, fmt.Errorf("not valid JSON: %w", err)
	}
	if err := expr.Validate(); err != nil {
		return queryir.Expr{}, err
	}
	return expr, nil
}

// stripFences removes a surrounding ```json ... ``` (or ``` ... ```) fence if the
// model wrapped its output in one.
func stripFences(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	// Drop an optional language tag on the first line.
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		firstLine := strings.TrimSpace(s[:i])
		if firstLine == "" || !strings.ContainsAny(firstLine, "{}\"") {
			s = s[i+1:]
		}
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}
