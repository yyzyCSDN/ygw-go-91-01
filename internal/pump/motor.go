package pump

import "sync"

type Starter interface {
	Start() error
	Stop() error
}

type Motor struct {
	mu      sync.Mutex
	running bool
}

func NewMotor() *Motor {
	return &Motor{}
}

func (m *Motor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

func (m *Motor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *Motor) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
