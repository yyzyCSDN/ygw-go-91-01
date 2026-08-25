package alarm

import "fuelhydrant/internal/model"

func (m *Manager) Reconcile(statuses []model.HydrantStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, status := range statuses {
		if !status.Leaking {
			continue
		}
		m.alarms[status.ID] = &Alarm{ID: status.ID, Kind: HydrantLeakKind(), Active: true}
	}
}
