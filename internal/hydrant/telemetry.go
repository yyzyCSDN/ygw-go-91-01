package hydrant

import "fuelhydrant/internal/model"

func (m *Monitor) ApplyTelemetry(t model.Telemetry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[t.Hydrant]
	if !ok {
		return
	}
	h.pressure = t.Pressure
	h.flow = t.Flow
	if !model.ValidPressure(t.Pressure) {
		return
	}
	m.cache[t.Hydrant] = t.Pressure
}

func (m *Monitor) ApplyTelemetryBatch(items []model.Telemetry) {
	for _, item := range items {
		m.ApplyTelemetry(item)
	}
}

func (m *Monitor) RefreshCache(id string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	if !ok || h == nil {
		return 0, false
	}
	m.cache[id] = h.pressure
	return h.pressure, true
}
