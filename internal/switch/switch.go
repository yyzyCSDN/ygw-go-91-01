package fuelswitch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fuelhydrant/internal/model"
)

type State struct {
	mu    sync.Mutex
	state model.SwitchState
	from  model.FuelType
	to    model.FuelType
}

func New() *State {
	return &State{state: model.SwitchIdle}
}

func (s *State) State() model.SwitchState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *State) From() model.FuelType {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.from
}

func (s *State) To() model.FuelType {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.to
}

func (s *State) InProgress() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == model.SwitchDraining || s.state == model.SwitchSwitching
}

func (s *State) Begin(from, to model.FuelType, drainConfirm func() bool, timeout time.Duration) error {
	s.mu.Lock()
	if s.state != model.SwitchIdle {
		s.mu.Unlock()
		return fmt.Errorf("fuel switch already in progress")
	}
	s.state = model.SwitchDraining
	s.from = from
	s.to = to
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			err := s.timeoutLocked()
			s.mu.Unlock()
			return err
		case <-ticker.C:
			if drainConfirm() {
				s.mu.Lock()
				s.state = model.SwitchSwitching
				s.mu.Unlock()
				return nil
			}
		}
	}
}

func (s *State) resetLocked() {
	s.state = model.SwitchIdle
}

func (s *State) timeoutLocked() error {
	s.resetLocked()
	s.from = ""
	s.to = ""
	return model.ErrSwitchTimeout
}

func (s *State) Complete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != model.SwitchSwitching {
		return fmt.Errorf("fuel switch not in switching state")
	}
	s.state = model.SwitchIdle
	return nil
}

func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = model.SwitchIdle
}
