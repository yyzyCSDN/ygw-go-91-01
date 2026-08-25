package leak

import "fuelhydrant/internal/model"

type Severity string

const (
	SeverityMinor    Severity = "minor"
	SeverityMajor    Severity = "major"
	SeverityCritical Severity = "critical"
)

func (d *Detector) ActiveLeaks() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	count := 0
	for _, active := range d.leaking {
		if active {
			count++
		}
	}
	return count
}

func (d *Detector) Severity(id string) Severity {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.leaking[id] {
		return ""
	}
	return SeverityCritical
}

func (d *Detector) PressureDrop(current, previous float64, threshold float64) bool {
	drop := previous - current
	return drop >= threshold && previous > 0
}

func (d *Detector) EvaluateStatus(statuses []model.HydrantStatus, threshold float64) bool {
	for _, status := range statuses {
		if status.Leaking {
			return true
		}
	}
	return false
}
