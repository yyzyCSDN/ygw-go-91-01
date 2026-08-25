package hydrant

import "fuelhydrant/internal/model"

type MonitorSnapshot struct {
	Hydrants []model.HydrantStatus
	Fuel     model.FuelType
}

func (m *Monitor) Snapshot() MonitorSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	hydrants := make([]model.HydrantStatus, 0, len(m.hydrants))
	for _, h := range m.hydrants {
		hydrants = append(hydrants, h.Status())
	}
	return MonitorSnapshot{Hydrants: hydrants}
}

func (m *Monitor) Recover(s MonitorSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(s.Hydrants) == 0 {
		return
	}
	m.rebuildLocked(s)
}

func (m *Monitor) rebuildLocked(s MonitorSnapshot) {
	for _, status := range s.Hydrants {
		h, ok := m.hydrants[status.ID]
		if !ok {
			h = &Hydrant{ID: status.ID, fuel: status.Fuel}
			m.hydrants[status.ID] = h
		}
		h.pressure = status.Pressure
		h.flow = status.Flow
		h.leaking = status.Leaking
		h.fuel = status.Fuel
		m.cache[status.ID] = status.Pressure
	}
}
