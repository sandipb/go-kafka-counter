package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

const defaultUpdateEveryMillis = 1000

func main() {
	var cfgPath string
	var debug bool
	var updateEveryMs int
	pflag.StringVarP(&cfgPath, "config", "c", "config.yaml", "Path to YAML config file")
	pflag.BoolVarP(&debug, "debug", "d", false, "Run in debug mode")
	pflag.IntVarP(&updateEveryMs, "update-every", "u", defaultUpdateEveryMillis, "How many ms between display updates")
	pflag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Msgf("Using config file: %s", cfgPath)

	appConfig := NewAppConfig(cfgPath)
	log.Debug().Msgf("Config:\n%s", appConfig.Encode())

	appConfig.Validate()
	kr := NewKafkaReader(appConfig.Common.Broker, appConfig.Topics())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	counter := NewEventCounter(appConfig.Topics(), updateEveryMs)
	go kr.DumpMessages(ctx, counter)
	<-signals
	log.Debug().Msg("Terminating")
	cancel()
	kr.Close()
}
