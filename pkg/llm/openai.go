package llm

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"
)

// openAICompleter is the production completer backed by go-openai.
type openAICompleter struct {
	client *openai.Client
	model  string
	temp   float32
}

func (o *openAICompleter) complete(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       o.model,
		Messages:    messages,
		Temperature: o.temp,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("llm: empty response from model")
	}
	return resp.Choices[0].Message.Content, nil
}
