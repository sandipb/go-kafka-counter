package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type EventCounter struct {
	Counts            map[string]map[string]int
	Total             int
	TopicTotals       map[string]int
	Topics            []string
	Config            *AppConfig
	UpdateEveryMillis int
	LastUpdate        time.Time
	NoProgress        bool
}

func NewEventCounter(appConfig *AppConfig, updateEveryMillis int, noProgress bool) *EventCounter {
	topics := appConfig.Topics()
	ec := EventCounter{
		Counts:            make(map[string]map[string]int, len(topics)),
		TopicTotals:       make(map[string]int, len(topics)),
		LastUpdate:        time.Now(),
		Config:            appConfig,
		Topics:            topics,
		UpdateEveryMillis: updateEveryMillis,
		NoProgress:        noProgress,
	}
	for _, topic := range topics {
		ec.Counts[topic] = map[string]int{}
		ec.TopicTotals[topic] = 0
	}
	return &ec
}

func (ec *EventCounter) Add(topic string, key string) {
	if _, ok := ec.Counts[topic][key]; !ok {
		ec.Counts[topic][key] = 0
	}
	ec.Counts[topic][key] += 1
	ec.Total += 1
	ec.TopicTotals[topic] += 1
	if !ec.NoProgress {
		ec.display(zerolog.GlobalLevel() != zerolog.DebugLevel)
	}
}

func (ec *EventCounter) display(clear bool) {
	if time.Since(ec.LastUpdate).Milliseconds() < int64(ec.UpdateEveryMillis) {
		return
	}
	topicLines := []string{}
	topicCounts := reverseSortCounts(ec.TopicTotals)
	for _, kv := range topicCounts {
		topic := kv.k
		topicTotal := kv.v
		name := ec.Config.Count[topic].Name
		if topicTotal == 0 {
			continue
		}
		sortedVals := reverseSortCounts(ec.Counts[topic])
		texts := make([]string, 0, len(sortedVals))
		for _, e := range sortedVals {
			texts = append(texts, fmt.Sprintf(
				"%s=%d(%.1f)%%", e.k, e.v, float64(e.v)*100/float64(topicTotal),
			))
		}
		topicLines = append(topicLines,
			fmt.Sprintf("%25s[%6d] (%s)", name, ec.TopicTotals[topic], strings.Join(texts, ", ")),
		)
	}
	if clear {
		os.Stdout.Write([]byte("\033c")) // clear
	}
	os.Stdout.Write([]byte(strings.Join(topicLines, "\n") + "\n"))
	// log.Info().Msgf("Log level is %v, debug=%v", zerolog.GlobalLevel(), zerolog.GlobalLevel() == zerolog.DebugLevel)
	ec.LastUpdate = time.Now()
}

func (ec *EventCounter) DisplayStats() {
	fmt.Println()
	if ec.Total == 0 {
		fmt.Println("No events received")
		return
	}
	fmt.Printf("Total events received: %d\n", ec.Total)
	topicCounts := reverseSortCounts(ec.TopicTotals)
	for _, entry := range topicCounts {
		name := ec.Config.Count[entry.k].Name
		if entry.v > 0 {
			fmt.Printf("\t%30s: %7d (%2.2f%%)\n", name, entry.v, float64(entry.v)*100/float64(ec.Total))
		}
	}
	fmt.Println("\nDetails:")
	ec.display(false)
}

type StringInt struct {
	k string
	v int
}

func reverseSortCounts(kv map[string]int) []StringInt {
	counts := []StringInt{}
	for k, v := range kv {
		counts = append(counts, StringInt{k, v})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].v > counts[j].v
	})
	return counts
}
