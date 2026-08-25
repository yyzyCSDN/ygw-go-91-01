package pump

import (
	"fmt"
	"sync"

	"fuelhydrant/internal/model"
)

type FailureNotifier interface {
	OnPumpStartFailure(id string, err error)
}

type Pump struct {
	id       string
	starter  Starter
	notifier FailureNotifier
	mu       sync.Mutex
	state    model.PumpState
	retries  int
}

func NewPump(id string, starter Starter, notifier FailureNotifier) *Pump {
	return &Pump{
		id:       id,
		starter:  starter,
		notifier: notifier,
		state:    model.PumpIdle,
	}
}

func (p *Pump) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.starter.Start(); err != nil {
		return p.startFailureLocked(err)
	}
	p.state = model.PumpRunning
	p.retries = 0
	return nil
}

func (p *Pump) startFailureLocked(err error) error {
	p.state = model.PumpFailed
	p.retries++
	if p.notifier != nil {
		p.notifier.OnPumpStartFailure(p.id, err)
	}
	return fmt.Errorf("pump %s start failed: %w", p.id, err)
}

func (p *Pump) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.starter.Stop(); err != nil {
		return fmt.Errorf("pump %s stop failed: %w", p.id, err)
	}
	p.state = model.PumpIdle
	return nil
}

func (p *Pump) State() model.PumpState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *Pump) Retries() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.retries
}

func (p *Pump) Failed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state == model.PumpFailed
}

func (p *Pump) ID() string {
	return p.id
}
