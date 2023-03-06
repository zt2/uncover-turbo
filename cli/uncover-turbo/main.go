package main

import (
	"context"
	"errors"
	"os"

	// Attempts to increase the OS file descriptors - Fail silently
	_ "github.com/projectdiscovery/fdmax/autofdmax"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/uncover/runner"
	"github.com/zt2/uncover-turbo/pkg/bot"
)

func main() {
	// Parse the command line flags and read config files
	options := runner.ParseOptions()

	newRunner, err := runner.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msgf("Could not create runner: %s\n", err)
	}

	if len(options.Query) > 0 {
		for _, query := range options.Query {
			if text, err := translate("fofa", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to fofa query: %s\n", text)

				options.Fofa = []string{text}
			}

			if text, err := translate("360 Quake", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to quake query: %s\n", text)

				options.Quake = []string{text}
			}

			if text, err := translate("zoomeye", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to zoomeye query: %s\n", text)

				options.ZoomEye = []string{text}
			}

			if text, err := translate("shodan", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to shodan query: %s\n", text)

				options.Shodan = []string{text}
			}
		}

		options.Query = []string{}
		options.Engine = []string{}
	}

	err = newRunner.Run(context.Background(), options.Query...)
	if err != nil {
		gologger.Fatal().Msgf("Could not run enumeration: %s\n", err)
	}
}

func translate(engine, query string) (string, error) {
	var b *bot.Bot

	if key, exists := os.LookupEnv("OPENAI_KEY"); !exists {
		return "", errors.New("no OpenAI key found")
	} else {
		b = bot.New(key)
	}

	return b.Translate(engine, query)
}
