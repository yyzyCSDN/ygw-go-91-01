package alarm

import (
	"sort"
	"sync"

	"fuelhydrant/internal/model"
)

type JournalWriter interface {
	Append(kind, body string) error
}

type RestoreTarget interface {
	Release(id string) error
}

type Alarm struct {
	ID     string
	Kind   string
	Active bool
}

type Manager struct {
	mu      sync.Mutex
	alarms  map[string]*Alarm
	journal JournalWriter
	restore RestoreTarget
}

func NewManager(journal JournalWriter, restore RestoreTarget) *Manager {
	return &Manager{
		alarms:  make(map[string]*Alarm),
		journal: journal,
		restore: restore,
	}
}

func (m *Manager) Raise(id, kind string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alarms[id] = &Alarm{ID: id, Kind: kind, Active: true}
	if m.journal != nil {
		_ = m.journal.Append("alarm", id+":"+kind)
	}
}

func (m *Manager) Clear(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	if !ok || !a.Active {
		return nil
	}
	if m.restore != nil {
		_ = m.restore.Release(id)
	}
	a.Active = false
	if m.journal != nil {
		_ = m.journal.Append("alarm-clear", id)
	}
	return nil
}



func (m *Manager) Active(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	return ok && a.Active
}

func (m *Manager) List() []Alarm {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.alarms))
	for id := range m.alarms {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Alarm, 0, len(ids))
	for _, id := range ids {
		out = append(out, *m.alarms[id])
	}
	return out
}

func (m *Manager) OnPumpStartFailure(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err == nil {
		return
	}
	m.alarms[id] = &Alarm{ID: id, Kind: "pump-start-failure", Active: true}
	if m.journal != nil {
		_ = m.journal.Append("alarm", id+":pump-start-failure:"+err.Error())
	}
}

func (m *Manager) ResolveHydrant(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	if ok && a.Kind == "hydrant-leak" {
		a.Active = false
	}
	if m.journal != nil {
		_ = m.journal.Append("alarm-resolve", id)
	}
}

func (m *Manager) RaiseHydrantLeak(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alarms[id] = &Alarm{ID: id, Kind: "hydrant-leak", Active: true}
}

func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, a := range m.alarms {
		if a.Active {
			count++
		}
	}
	return count
}

func (m *Manager) Kind(id string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alarms[id]
	if !ok {
		return ""
	}
	return a.Kind
}

func HydrantLeakKind() string {
	return "hydrant-leak"
}

func (m *Manager) RestoreTarget() RestoreTarget {
	return m.restore
}

func (m *Manager) RecordEvent(kind, body string) error {
	if m.journal == nil {
		return model.ErrRestoreFailed
	}
	return m.journal.Append(kind, body)
}
