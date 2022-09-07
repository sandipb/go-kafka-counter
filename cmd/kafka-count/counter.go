package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type EventCounter struct {
	Counts            map[string]map[string]int
	Total             int
	TopicTotals       map[string]int
	Topics            []string
	UpdateEveryMillis int
	LastUpdate        time.Time
}

func NewEventCounter(topics []string, updateEveryMillis int) *EventCounter {
	ec := EventCounter{
		Counts:            make(map[string]map[string]int, len(topics)),
		TopicTotals:       make(map[string]int, len(topics)),
		Topics:            topics,
		UpdateEveryMillis: updateEveryMillis,
		LastUpdate:        time.Now(),
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
	ec.display()
}

func (ec *EventCounter) display() {
	if time.Since(ec.LastUpdate).Milliseconds() < int64(ec.UpdateEveryMillis) {
		return
	}
	topicLines := []string{}
	for _, topic := range ec.Topics {
		if ec.TopicTotals[topic] == 0 {
			continue
		}
		sortedVals := reverseSortCounts(ec.Counts[topic])
		texts := make([]string, 0, len(sortedVals))
		for _, e := range sortedVals {
			texts = append(texts, fmt.Sprintf(
				"%s=%d(%.1f)%%", e.k, e.v, float64(e.v)*100/float64(ec.TopicTotals[topic]),
			))
		}
		topicLines = append(topicLines,
			fmt.Sprintf("%25s[%d](%s)", topic, ec.TopicTotals[topic], strings.Join(texts, ", ")),
		)
	}
	os.Stdout.Write([]byte("\033c"))                               // clear
	os.Stdout.Write([]byte(strings.Join(topicLines, "\n") + "\n")) // clear
	ec.LastUpdate = time.Now()
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
