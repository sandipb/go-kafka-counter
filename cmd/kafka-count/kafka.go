package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type KafkaReader struct {
	Config   *AppConfig
	Logger   zerolog.Logger
	Consumer *kafka.Consumer
}

func NewKafkaReader(broker string, appConfig *AppConfig) *KafkaReader {
	logger := log.With().Str("ctx", "kafka").Logger()
	kr := KafkaReader{Config: appConfig, Logger: logger}

	logger.Debug().Msgf("Using brokers: %v, and reading topics %v", broker, appConfig.Topics())
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
	kr.Consumer.SubscribeTopics(appConfig.Topics(), nil)
	return &kr
}

func (kr *KafkaReader) DumpMessages(ctx context.Context, stopChan chan struct{}, ec *EventCounter, maxCount int) {
	logger := kr.Logger
	topics := kr.Config.Topics()
	topicMap := map[string]struct{}{}
	for _, t := range topics {
		topicMap[t] = struct{}{}
	}

	logger.Debug().Msg("Beginning DumpMessages")
	count := 0

	for {
		select {
		case <-ctx.Done():
			logger.Debug().Msg("Detected context cancellation. Exiting DumpMessages.")
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
			if _, ok := topicMap[*msg.TopicPartition.Topic]; !ok {
				logger.Warn().Msgf("Ignoring unexpected topic: %s", *msg.TopicPartition.Topic)
				continue
			}
			kr.ProcessMessage(msg, ec)
			count += 1
			if maxCount > 0 && count == maxCount {
				close(stopChan)
				time.Sleep(10 * time.Millisecond) // give a chance to main thread to process the signal
			}
		}
	}
}

func (kr *KafkaReader) ProcessMessage(msg *kafka.Message, ec *EventCounter) {
	msgTopic := *msg.TopicPartition.Topic
	payloadRaw := string(msg.Value)
	topicConfig := kr.Config.Count[msgTopic]

	if topicConfig.CountOnly {
		ec.Add(msgTopic, "COUNT")
		return
	}

	kr.Logger.Debug().Msgf("%s: %s", msgTopic, payloadRaw)

	var payload interface{}
	if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
		ec.Add(msgTopic, "NON-JSON")
		return
	}

	result, err := topicConfig.PathCompiled.Search(payload)
	if err != nil {
		ec.Add(msgTopic, "SEARCH-ERROR")
		kr.Logger.Error().Err(err).Msgf("Search error for payload: %s", payloadRaw)
		return
	}

	if resultStr, ok := result.(string); !ok {
		kr.Logger.Error().Msgf("Got non-string key %v (%T)", result, result)
		ec.Add(msgTopic, "NON-STRING-KEY")
	} else if resultStr == "" {
		ec.Add(msgTopic, "NON-MATCH")
	} else {
		ec.Add(msgTopic, resultStr)
	}
}

func (kr *KafkaReader) Close() error {
	kr.Logger.Debug().Msg("Closing reader")
	return kr.Consumer.Close()
}
