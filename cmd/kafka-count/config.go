package main

import (
	"bytes"
	"os"

	"github.com/jmespath/go-jmespath"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Common struct {
		Broker string `yaml:"broker"`
	} `yaml:"common"`
	Count map[string]struct {
		Name         string             `yaml:"name"`
		Path         string             `yaml:"path"`
		PathCompiled *jmespath.JMESPath `yaml:"-"`
		CountOnly    bool               `yaml:"countOnly"`
	} `yaml:"count"`
}

func NewAppConfig(cfgPath string) *AppConfig {
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not read config file")
	}
	var config AppConfig
	if err := yaml.NewDecoder(bytes.NewReader(raw)).Decode(&config); err != nil {
		log.Fatal().Err(err).Msg("Could not parse config file")
	}
	return &config
}

func (ac *AppConfig) Validate() {
	logger := log.With().Str("ctx", "cfgValidate").Logger()
	if ac.Common.Broker == "" {
		logger.Fatal().Msg("Broker not specified")
	}
	for topic, entry := range ac.Count {
		if entry.Name == "" {
			entry.Name = topic
		}
		ac.Count[topic] = entry
		if entry.CountOnly {
			continue
		}
		if entry.Path == "" {
			logger.Fatal().Msgf("No path specified for topic %q", topic)
		}

		if expr, err := jmespath.Compile(entry.Path); err != nil {
			logger.Fatal().Msgf("Could not parse path %s for topic %q", entry.Path, topic)
		} else {
			entry.PathCompiled = expr
			ac.Count[topic] = entry
		}
	}
}

func (ac *AppConfig) Topics() []string {
	topics := make([]string, 0, len(ac.Count))
	for topic := range ac.Count {
		topics = append(topics, topic)
	}
	return topics
}

func (ac *AppConfig) Encode() string {
	b := bytes.NewBuffer(nil)
	if err := yaml.NewEncoder(b).Encode(ac); err != nil {
		log.Fatal().Err(err).Msg("Could not encode config")
	}
	return string(b.Bytes())
}
