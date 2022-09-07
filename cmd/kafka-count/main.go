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
	var noProgress bool
	var updateEveryMs int
	var maxCount int
	pflag.StringVarP(&cfgPath, "config", "c", "config.yaml", "Path to YAML config file")
	pflag.BoolVarP(&debug, "debug", "d", false, "Run in debug mode")
	pflag.BoolVarP(&noProgress, "no-progress", "N", false, "Do not show running progress. Only the stats at the end.")
	pflag.IntVarP(&updateEveryMs, "update-every", "u", defaultUpdateEveryMillis, "How many ms between display updates")
	pflag.IntVarP(&maxCount, "max-count", "n", -1, "How many messages to process before exiting. < 0 means infinite.")
	pflag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// log.Debug().Msgf("Using config file: %s", cfgPath)

	appConfig := NewAppConfig(cfgPath)
	log.Debug().Msgf("Config:\n%s", appConfig.Encode())

	appConfig.Validate()
	kr := NewKafkaReader(appConfig.Common.Broker, appConfig)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	counter := NewEventCounter(appConfig, updateEveryMs, noProgress)

	// channel for kafka reader to let us know it is done.
	stopChan := make(chan struct{})
	go kr.DumpMessages(ctx, stopChan, counter, maxCount)

	select {
	case <-signals:
	case <-stopChan:
	}

	log.Debug().Msg("Terminating")
	cancel()
	kr.Close()
	counter.DisplayStats()
}
