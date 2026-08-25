package hydrant

import "fuelhydrant/internal/model"

func (m *Monitor) AveragePressure() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.hydrants) == 0 {
		return 0
	}
	values := make([]float64, 0, len(m.hydrants))
	for _, h := range m.hydrants {
		values = append(values, h.pressure)
	}
	return model.Average(values)
}

func (m *Monitor) TotalFlow() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	values := make([]float64, 0, len(m.hydrants))
	for _, h := range m.hydrants {
		values = append(values, h.flow)
	}
	return model.Sum(values)
}

func (m *Monitor) LeakingCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, h := range m.hydrants {
		if h.leaking {
			count++
		}
	}
	return count
}

func (m *Monitor) FlowRate(id string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	if !ok || h == nil {
		return 0, false
	}
	return h.flow, true
}

func (m *Monitor) SetFuel(id string, fuel string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	if !ok {
		return false
	}
	h.fuel = modelFuelType(fuel)
	return true
}
