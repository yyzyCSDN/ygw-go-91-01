package fuelswitch

import (
	"sync"

	"fuelhydrant/internal/model"
)

type FlushTracker struct {
	mu       sync.Mutex
	cycles   int
	lastFuel model.FuelType
}

func NewFlushTracker() *FlushTracker {
	return &FlushTracker{}
}

func (f *FlushTracker) RecordCycle(fuel model.FuelType) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cycles++
	f.lastFuel = fuel
}

func (f *FlushTracker) Cycles() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.cycles
}

func (f *FlushTracker) LastFuel() model.FuelType {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastFuel
}
