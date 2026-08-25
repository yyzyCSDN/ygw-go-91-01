package supply

import (
	"fmt"
	"sync"
	"time"

	"fuelhydrant/internal/hydrant"
	"fuelhydrant/internal/model"
	"fuelhydrant/internal/pump"
	"fuelhydrant/internal/switch"
)

type Coordinator struct {
	mu        sync.Mutex
	state     model.SupplyState
	fuel      model.FuelType
	target    float64
	tolerance float64
	monitor   *hydrant.Monitor
	group     *pump.Group
	switcher  *fuelswitch.State
	drain     *fuelswitch.Drain
	flush     *fuelswitch.FlushTracker
}

func NewCoordinator(monitor *hydrant.Monitor, group *pump.Group, switcher *fuelswitch.State) *Coordinator {
	return &Coordinator{
		state:     model.SupplyIdle,
		fuel:      model.FuelJetA1,
		target:    4.0,
		tolerance: 0.2,
		monitor:   monitor,
		group:     group,
		switcher:  switcher,
		drain:     fuelswitch.NewDrain(model.FuelJetA1),
		flush:     fuelswitch.NewFlushTracker(),
	}
}

func (c *Coordinator) State() model.SupplyState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *Coordinator) Fuel() model.FuelType {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fuel
}

func (c *Coordinator) SetTarget(target, tolerance float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.target = target
	c.tolerance = tolerance
}

func (c *Coordinator) HydrantPressure(id string) (float64, error) {
	h := c.monitor.HydrantPtr(id)
	return h.PressureValue(), nil
}

func (c *Coordinator) Start() error {
	c.mu.Lock()
	if c.state != model.SupplyIdle {
		c.mu.Unlock()
		return fmt.Errorf("supply not idle: %s", c.state)
	}
	c.state = model.SupplyPressurizing
	c.mu.Unlock()

	if err := c.WaitPressureStable(30 * time.Second); err != nil {
		c.mu.Lock()
		c.state = model.SupplyIdle
		c.mu.Unlock()
		return err
	}
	if !c.AutoAllowed() {
		c.mu.Lock()
		c.state = model.SupplyIdle
		c.mu.Unlock()
		return nil
	}
	if err := c.group.StartAll(); err != nil {
		c.mu.Lock()
		c.state = model.SupplyIdle
		c.mu.Unlock()
		return err
	}
	c.mu.Lock()
	c.state = model.SupplySupplying
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) Stop() error {
	c.mu.Lock()
	if c.state != model.SupplySupplying {
		c.mu.Unlock()
		return fmt.Errorf("supply not supplying: %s", c.state)
	}
	c.state = model.SupplyStopping
	c.mu.Unlock()

	if err := c.group.StopAll(); err != nil {
		c.mu.Lock()
		c.state = model.SupplySupplying
		c.mu.Unlock()
		return err
	}
	c.mu.Lock()
	c.state = model.SupplyIdle
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) SwitchFuel(target model.FuelType) error {
	c.mu.Lock()
	if c.state != model.SupplyIdle && c.state != model.SupplyStopping {
		c.mu.Unlock()
		return fmt.Errorf("cannot switch fuel while %s", c.state)
	}
	current := c.fuel
	c.mu.Unlock()

	if !target.Valid() {
		return fmt.Errorf("invalid fuel type %q", target)
	}
	if err := c.switcher.Begin(current, target, c.drainDone, 15*time.Second); err != nil {
		return fmt.Errorf("switch %s -> %s: %w", current, target, err)
	}
	c.dispatchFuelLocked(target)
	if err := c.switcher.Complete(); err != nil {
		return fmt.Errorf("complete switch %s -> %s: %w", current, target, err)
	}
	c.flush.RecordCycle(target)
	return nil
}

func (c *Coordinator) dispatchFuelLocked(target model.FuelType) {
	c.group.ApplyFuelSequence(target)
	c.mu.Lock()
	c.fuel = target
	c.mu.Unlock()
}

func (c *Coordinator) AutoAllowed() bool {
	return !c.group.ManualState().Active
}

func (c *Coordinator) drainDone() bool {
	return c.drain.Done()
}

func (c *Coordinator) DispatchSequence(fuel model.FuelType) {
	c.group.SetFuel(fuel)
	c.mu.Lock()
	c.fuel = fuel
	c.mu.Unlock()
}

func (c *Coordinator) SetDrainProgress(progress float64) {
	c.drain.SetProgress(model.Clamp(progress, 0, 1))
}

func (c *Coordinator) FlushCycles() int {
	return c.flush.Cycles()
}

func (c *Coordinator) SwitchStatus() map[string]any {
	return map[string]any{
		"state": c.switcher.State(),
		"from":  c.switcher.From(),
		"to":    c.switcher.To(),
		"drain": c.drain.Progress(),
	}
}

func (c *Coordinator) LastFuel() model.FuelType {
	return c.flush.LastFuel()
}
