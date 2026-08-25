package record

import (
	"bufio"
	"os"
	"strings"

	"fuelhydrant/internal/model"
)

func (j *Journal) ReadEvents() ([]model.RecordEvent, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if err := j.flushLocked(); err != nil {
		return nil, err
	}
	f, err := os.Open(j.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.RecordEvent{}, nil
		}
		return nil, err
	}
	defer f.Close()
	var events []model.RecordEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		events = append(events, model.RecordEvent{Kind: parts[0], Body: parts[1]})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (j *Journal) flushLocked() error {
	if j.writer == nil {
		return ErrNotOpen
	}
	return j.writer.Flush()
}
