package record

import "fuelhydrant/internal/model"

func (j *Journal) Page(offset, limit int) ([]model.RecordEvent, error) {
	events, err := j.ReadEvents()
	if err != nil {
		return nil, err
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(events) {
		return []model.RecordEvent{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(events) {
		end = len(events)
	}
	return events[offset:end], nil
}
