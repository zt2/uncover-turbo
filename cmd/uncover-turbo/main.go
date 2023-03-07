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

	if len(options.Censys) == 0 &&
		len(options.Quake) == 0 &&
		len(options.Fofa) == 0 {
		gologger.Fatal().Msg("no engine specified, only censys/quake/fofa supported for now, plese use -fofa/-quake/-censys instead of -query")
	}

	if len(options.Censys) > 0 {
		for idx, query := range options.Censys {
			if text, err := translate("censys", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to censys query: \"%s\"\n", text)

				options.Censys[idx] = text
			}
		}
	}

	if len(options.Fofa) > 0 {
		for idx, query := range options.Fofa {
			if text, err := translate("fofa", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to fofa query: \"%s\"\n", text)

				options.Fofa[idx] = text
			}
		}
	}

	if len(options.Quake) > 0 {
		for idx, query := range options.Quake {
			if text, err := translate("quake", query); err != nil {
				gologger.Fatal().Msgf("Could not translate query: %s\n", err)
			} else {
				gologger.Debug().Msgf("Translate to quake query: \"%s\"\n", text)

				options.Quake[idx] = text
			}
		}
	}

	// Disable default parameters
	options.Query = []string{}
	options.Engine = []string{}

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
