package fuelswitch

import (
	"sync"

	"fuelhydrant/internal/model"
)

type Drain struct {
	mu       sync.Mutex
	fuel     model.FuelType
	progress float64
}

func NewDrain(fuel model.FuelType) *Drain {
	return &Drain{fuel: fuel}
}

func (d *Drain) SetProgress(progress float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if progress > 1 {
		progress = 1
	}
	if progress < 0 {
		progress = 0
	}
	d.progress = progress
}

func (d *Drain) Progress() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.progress
}

func (d *Drain) Done() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.progress >= 1
}

func (d *Drain) Fuel() model.FuelType {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.fuel
}
