package alarm

import "fuelhydrant/internal/model"

func (m *Manager) Reconcile(statuses []model.HydrantStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	leaking := make(map[string]bool)
	for _, status := range statuses {
		leaking[status.ID] = status.Leaking
	}
	for id, a := range m.alarms {
		if a.Kind != HydrantLeakKind() {
			continue
		}
		isLeaking, known := leaking[id]
		if !known || !isLeaking {
			a.Active = false
			if m.journal != nil {
				_ = m.journal.Append("alarm-resolve", id)
			}
		}
	}
	for _, status := range statuses {
		if !status.Leaking {
			continue
		}
		m.alarms[status.ID] = &Alarm{ID: status.ID, Kind: HydrantLeakKind(), Active: true}
	}
}
