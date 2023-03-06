package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/projectdiscovery/uncover/runner"
	"github.com/sashabaranov/go-openai"
)

func New(authToken string) *Bot {
	return &Bot{
		Client: openai.NewClient(authToken),
	}
}

type Bot struct {
	Client *openai.Client
}

func (bot *Bot) Translate(engine, text string) (string, error) {
	initialMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(systemPrompt, engine),
	}

	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	}

	resp, err := bot.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    []openai.ChatCompletionMessage{initialMessage, message},
			Temperature: 0.0,
		},
	)

	if err != nil {
		return "", nil
	}

	content := resp.Choices[0].Message.Content
	return strings.Trim(content, "\r\n\t "), nil
}

func (bot *Bot) Query(query string) error {
	options := &runner.Options{}
	r, err := runner.NewRunner(options)
	if err != nil {
		return err
	}

	return r.Run(context.Background(), query)
}
