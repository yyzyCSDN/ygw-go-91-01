package supply

import (
	"sort"

	"fuelhydrant/internal/model"
)

type Sequence struct {
	Fuel    model.FuelType
	PumpIDs []string
}

func (c *Coordinator) Sequence() Sequence {
	c.mu.Lock()
	defer c.mu.Unlock()
	ids := c.group.IDs()
	sort.Strings(ids)
	return Sequence{Fuel: c.fuel, PumpIDs: ids}
}

func (c *Coordinator) ApplySequence(seq Sequence) {
	c.group.SetFuel(seq.Fuel)
	c.mu.Lock()
	c.fuel = seq.Fuel
	c.mu.Unlock()
}
