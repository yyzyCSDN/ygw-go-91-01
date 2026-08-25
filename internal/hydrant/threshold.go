package hydrant

import "fuelhydrant/internal/model"

type Thresholds struct {
	LowPressure  float64
	HighPressure float64
	LowFlow      float64
}

func DefaultThresholds() Thresholds {
	return Thresholds{LowPressure: 2.5, HighPressure: 6.0, LowFlow: 0.5}
}

func (m *Monitor) BelowPressure(threshold float64) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0)
	for _, h := range m.hydrants {
		if h.pressure < threshold {
			ids = append(ids, h.ID)
		}
	}
	return ids
}

func (m *Monitor) Evaluate(id string, thresholds Thresholds) (lowPressure, lowFlow bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hydrants[id]
	if !ok {
		return false, false
	}
	return h.pressure < thresholds.LowPressure, h.flow < thresholds.LowFlow
}

func modelFuelType(value string) model.FuelType {
	return model.FuelType(value)
}
