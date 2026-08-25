package leak

import (
	"sync"

	"fuelhydrant/internal/model"
)

type Detector struct {
	mu         sync.Mutex
	interlock  *Interlock
	raiseAlarm func(id string)
	stopPumps  func() error
	writeback  func(id string) error
	leaking    map[string]bool
}

func NewDetector() *Detector {
	return &Detector{
		interlock: NewInterlock(),
		leaking:   make(map[string]bool),
	}
}

func (d *Detector) SetRaiseAlarm(fn func(id string)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.raiseAlarm = fn
}

func (d *Detector) SetStopPumps(fn func() error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopPumps = fn
}

func (d *Detector) SetWriteback(fn func(id string) error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.writeback = fn
}

func (d *Detector) InterlockState() model.InterlockState {
	return d.interlock.State()
}

func (d *Detector) Detect(statuses []model.HydrantStatus) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	found := false
	for _, status := range statuses {
		if status.Leaking {
			found = true
			d.leaking[status.ID] = true
			if d.raiseAlarm != nil {
				d.raiseAlarm(status.ID)
			}
		}
	}
	if found {
		_ = d.interlock.Lock()
		if d.stopPumps != nil {
			_ = d.stopPumps()
		}
	}
	return found
}

func (d *Detector) Release(id string) error {
	return nil
}

func (d *Detector) Leaking(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.leaking[id]
}

func (d *Detector) ResetInterlock() {
	d.interlock.Reset()
}
