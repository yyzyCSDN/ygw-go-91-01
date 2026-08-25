package hydrant

import (
	"sort"
	"sync"

	"fuelhydrant/internal/model"
)

type Hydrant struct {
	ID       string
	pressure float64
	flow     float64
	leaking  bool
	fuel     model.FuelType
}

func (h *Hydrant) PressureValue() float64 {
	if h == nil {
		return 0
	}
	return h.pressure
}

func (h *Hydrant) FlowValue() float64 {
	if h == nil {
		return 0
	}
	return h.flow
}

func (h *Hydrant) Leaking() bool {
	if h == nil {
		return false
	}
	return h.leaking
}

func (h *Hydrant) FuelType() model.FuelType {
	if h == nil {
		return ""
	}
	return h.fuel
}

func (h *Hydrant) Status() model.HydrantStatus {
	if h == nil {
		return model.HydrantStatus{}
	}
	return model.HydrantStatus{
		ID:       h.ID,
		Pressure: h.pressure,
		Flow:     h.flow,
		Leaking:  h.leaking,
		Fuel:     h.fuel,
	}
}

type Monitor struct {
	mu           sync.Mutex
	hydrants     map[string]*Hydrant
	cache        map[string]float64
	lastSnapshot MonitorSnapshot
}

func NewMonitor() *Monitor {
	return &Monitor{
		hydrants: make(map[string]*Hydrant),
		cache:    make(map[string]float64),
	}
}

func (m *Monitor) Register(id string, fuel model.FuelType) *Hydrant {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := &Hydrant{ID: id, fuel: fuel}
	m.hydrants[id] = h
	m.cache[id] = h.pressure
	return h
}

func (m *Monitor) HydrantPtr(id string) *Hydrant {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hydrants[id]
}

func (m *Monitor) PressureOf(id string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := m.hydrants[id]
	return h.pressure, true
}

func (m *Monitor) Status(id string) (model.HydrantStatus, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	if !ok || h == nil {
		return model.HydrantStatus{}, false
	}
	return h.Status(), true
}

func (m *Monitor) Pressure(id string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.cache[id]
	return p, ok
}

func (m *Monitor) SetLeaking(id string, leaking bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	if !ok {
		return
	}
	h.leaking = leaking
}

func (m *Monitor) List() []model.HydrantStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.hydrants))
	for id := range m.hydrants {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]model.HydrantStatus, 0, len(ids))
	for _, id := range ids {
		out = append(out, m.hydrants[id].Status())
	}
	return out
}

func (m *Monitor) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.hydrants)
}

func (m *Monitor) Known(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	return ok && h != nil
}

func (m *Monitor) CacheSnapshot() {
	m.mu.Lock()
	defer m.mu.Unlock()
	hydrants := make([]model.HydrantStatus, 0, len(m.hydrants))
	for _, h := range m.hydrants {
		hydrants = append(hydrants, h.Status())
	}
	m.lastSnapshot = MonitorSnapshot{Hydrants: hydrants}
}
