package main

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type KafkaReader struct {
	Brokers  string
	Logger   zerolog.Logger
	Consumer *kafka.Consumer
}

func NewKafkaReader(broker string, topics []string) *KafkaReader {
	logger := log.With().Str("ctx", "kafka").Logger()
	kr := KafkaReader{Brokers: broker, Logger: logger}

	logger.Debug().Msgf("Using brokers: %v, and reading topics %v", broker, topics)
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"group.id":           "kafka-count",
		"auto.offset.reset":  "latest",
		"enable.auto.commit": false,
	})

	if err != nil {
		logger.Fatal().Err(err).Msg("Could not create consumer")
	}
	kr.Consumer = consumer
	kr.Consumer.SubscribeTopics(topics, nil)
	return &kr
}

func (kr *KafkaReader) DumpMessages(ctx context.Context, ec *EventCounter) {
	logger := kr.Logger
	logger.Debug().Msg("Beginning DumpMessages")
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Detected context cancellation. Exiting DumpMessages.")
			return
		default:
			msg, err := kr.Consumer.ReadMessage(500 * time.Millisecond)
			if err != nil {
				if kerror, ok := err.(kafka.Error); ok && kerror.Code() == kafka.ErrTimedOut {
					// ignore
				} else {
					logger.Error().Err(err).Msg("Reader had an error")
				}
				continue
			}
			ec.Add(*msg.TopicPartition.Topic, "COUNT")
			logger.Debug().Msgf("%s: %s", *msg.TopicPartition.Topic, string(msg.Value))
		}
	}
}

func (kr *KafkaReader) Close() error {
	kr.Logger.Debug().Msg("Closing reader")
	return kr.Consumer.Close()
}
