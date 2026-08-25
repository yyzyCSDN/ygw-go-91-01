package alarm

func (m *Manager) Acknowledge(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	if !ok {
		return false
	}
	a.Active = false
	if m.journal != nil {
		_ = m.journal.Append("alarm-ack", id)
	}
	return true
}

func (m *Manager) Acknowledged(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	if !ok {
		return false
	}
	return !a.Active
}
