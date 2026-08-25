package record

import (
	"sort"
	"strings"

	"fuelhydrant/internal/model"
)

func (j *Journal) QueryEvents(kind string) ([]model.RecordEvent, error) {
	events, err := j.ReadEvents()
	if err != nil {
		return nil, err
	}
	out := make([]model.RecordEvent, 0, len(events))
	for _, event := range events {
		if kind == "" || event.Kind == kind {
			out = append(out, event)
		}
	}
	return out, nil
}

func (j *Journal) CountByKind() (map[string]int, error) {
	events, err := j.ReadEvents()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, event := range events {
		counts[event.Kind]++
	}
	return counts, nil
}

func (j *Journal) Export() (string, error) {
	events, err := j.ReadEvents()
	if err != nil {
		return "", err
	}
	lines := make([]string, 0, len(events))
	for _, event := range events {
		lines = append(lines, event.Kind+"\t"+event.Body)
	}
	return strings.Join(lines, "\n"), nil
}

func (j *Journal) ChecksumOf(kind string) (string, error) {
	events, err := j.QueryEvents(kind)
	if err != nil {
		return "", err
	}
	return EventChecksum(events), nil
}

func SortedKinds(counts map[string]int) []string {
	kinds := make([]string, 0, len(counts))
	for kind := range counts {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}
