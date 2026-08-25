package pump

import (
	"sort"
	"sync"

	"fuelhydrant/internal/model"
)

type Group struct {
	mu       sync.Mutex
	pumps    map[string]*Pump
	mode     model.PumpMode
	fuel     model.FuelType
	pressure float64
	running  map[string]bool
	manual   map[string]bool
	hydrantPressure map[string]float64
}



func NewGroup() *Group {
	return &Group{
		pumps:   make(map[string]*Pump),
		mode:    model.PumpModeAuto,
		running: make(map[string]bool),
		manual:  make(map[string]bool),
		hydrantPressure: make(map[string]float64),
	}
}

func (g *Group) AddPump(p *Pump) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pumps[p.ID()] = p
}

func (g *Group) Pump(id string) (*Pump, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.pumps[id]
	return p, ok
}

func (g *Group) IDs() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	ids := make([]string, 0, len(g.pumps))
	for id := range g.pumps {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (g *Group) SetMode(mode model.PumpMode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.mode = mode
}

func (g *Group) Mode() model.PumpMode {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.mode
}

func (g *Group) SetFuel(f model.FuelType) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.fuel = f
}

func (g *Group) IsManual() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.mode == model.PumpModeManual
}



// ApplyFuelSequence records the fuel type for the group and clears the running
// set so the automatic sequence restarts cleanly. It never touches pumps the
// operator is running manually, so a fuel re-dispatch cannot silently drop a
// hand-started pump's state.
func (g *Group) ApplyFuelSequence(f model.FuelType) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !f.Valid() {
		return
	}
	g.fuel = f
	for id := range g.running {
		if g.manual[id] {
			continue
		}
		g.running[id] = false
	}
	if !g.anyManualLocked() {
		g.pressure = 0
	}
}

func (g *Group) Fuel() model.FuelType {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.fuel
}

func (g *Group) SetPressure(pressure float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pressure = pressure
}

func (g *Group) SyncHydrantPressure(id string, pressure float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if id == "" {
		return
	}
	g.hydrantPressure[id] = pressure
	g.recomputePressureLocked()
}

func (g *Group) recomputePressureLocked() {
	total := 0.0
	count := 0
	for _, value := range g.hydrantPressure {
		if value < 0 {
			continue
		}
		total += value
		count++
	}
	if count > 0 {
		g.pressure = total / float64(count)
		return
	}
	g.pressure = 0
}

func (g *Group) HydrantPressure(id string) (float64, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	value, ok := g.hydrantPressure[id]
	return value, ok
}

func (g *Group) Pressure() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.pressure
}

func (g *Group) PressureStable(target, tolerance float64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.pressureDeltaLocked(target) <= tolerance
}

func (g *Group) pressureDeltaLocked(target float64) float64 {
	return model.AbsDelta(g.pressure, target)
}

// ManualStart starts a single pump on operator demand. It flips the group into
// manual mode and marks the pump as manually controlled so the automatic supply
// logic does not overwrite (stop or restart) the operator's intent.
func (g *Group) ManualStart(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.pumps[id]
	if !ok {
		return model.ErrHydrantNotFound
	}
	if err := p.Start(); err != nil {
		return err
	}
	g.running[id] = true
	g.manual[id] = true
	g.mode = model.PumpModeManual
	return nil
}

// ManualStop stops a single pump on operator demand. It clears the pump's manual
// flag; when no pump remains under manual control the group returns to auto so
// the automatic supply logic can resume.
func (g *Group) ManualStop(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.pumps[id]
	if !ok {
		return model.ErrHydrantNotFound
	}
	if err := p.Stop(); err != nil {
		return err
	}
	g.running[id] = false
	g.manual[id] = false
	if !g.anyManualLocked() {
		g.mode = model.PumpModeAuto
	}
	return nil
}

// anyManualLocked reports whether any pump is still under operator manual
// control. Caller must hold g.mu.
func (g *Group) anyManualLocked() bool {
	for _, m := range g.manual {
		if m {
			return true
		}
	}
	return false
}

// StartAll starts every pump that is not already running. It yields to the
// operator: while the group is in manual mode (a pump is under manual control)
// it does nothing, so the automatic supply logic cannot overwrite a freshly
// hand-started pump.
func (g *Group) StartAll() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mode == model.PumpModeManual {
		return nil
	}
	for id, p := range g.pumps {
		if g.running[id] {
			continue
		}
		if err := p.Start(); err != nil {
			return err
		}
		g.running[id] = true
	}
	return nil
}

// StopAll stops the pumps driven by the automatic supply logic. It leaves any
// pump the operator started manually untouched, so the automatic stop path
// does not override a hand-started pump. Use StopAllForced for safety
// interlocks that must halt every pump regardless of operator intent.
func (g *Group) StopAll() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	var first error
	for id, p := range g.pumps {
		if g.manual[id] {
			continue
		}
		if err := p.Stop(); err != nil && first == nil {
			first = err
		}
		g.running[id] = false
	}
	return first
}

// StopAllForced halts every pump including those under manual control. It is
// the emergency path used by safety interlocks (e.g. leak shutoff) where
// stopping all pumps must not be deferred to operator action.
func (g *Group) StopAllForced() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	var first error
	for id, p := range g.pumps {
		if err := p.Stop(); err != nil && first == nil {
			first = err
		}
		g.running[id] = false
		g.manual[id] = false
	}
	if !g.anyManualLocked() {
		g.mode = model.PumpModeAuto
	}
	return first
}

func (g *Group) IsRunning() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, p := range g.pumps {
		if p.State() == model.PumpRunning {
			return true
		}
	}
	return false
}

func (g *Group) State(id string) model.PumpState {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.pumps[id]
	if !ok {
		return model.PumpIdle
	}
	return p.State()
}

func (g *Group) MotorRunning(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.pumps[id]
	if !ok {
		return false
	}
	motor, ok := p.starter.(*Motor)
	if !ok {
		return false
	}
	return motor.Running()
}
