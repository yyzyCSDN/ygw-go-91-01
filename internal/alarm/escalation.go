package alarm

type Escalation struct {
	ID    string
	Level int
}

func (m *Manager) Escalate(id string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	if !ok {
		return 0
	}
	a.Kind = "escalated:" + a.Kind
	return 1
}

func (m *Manager) ActiveAlarms() []Alarm {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Alarm, 0)
	for _, a := range m.alarms {
		if a.Active {
			out = append(out, *a)
		}
	}
	return out
}

func (m *Manager) History() []Alarm {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Alarm, 0, len(m.alarms))
	for _, a := range m.alarms {
		out = append(out, *a)
	}
	return out
}

func (m *Manager) TotalAlarms() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.alarms)
}
