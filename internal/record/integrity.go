package record

import (
	"encoding/hex"

	"github.com/cespare/xxhash/v2"

	"fuelhydrant/internal/model"
)

type Digest struct {
	Events int
	Sum    string
}

func (j *Journal) Integrity() (Digest, error) {
	events, err := j.ReadEvents()
	if err != nil {
		return Digest{}, err
	}
	return Digest{Events: len(events), Sum: EventChecksum(events)}, nil
}

func EventChecksum(events []model.RecordEvent) string {
	hasher := xxhash.New()
	for _, event := range events {
		_, _ = hasher.WriteString(event.Kind)
		_, _ = hasher.WriteString(":")
		_, _ = hasher.WriteString(event.Body)
		_, _ = hasher.WriteString("\n")
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
