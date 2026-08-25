package fuelswitch

import (
	"testing"
	"time"

	"fuelhydrant/internal/model"
)

// neverConfirmed never reports the drain as done, so Begin can only exit via
// its timeout. This reproduces the "排空等待超时" scenario from the bug
// report.
func neverConfirmed() bool { return false }

func TestBeginTimeoutResetsToIdle(t *testing.T) {
	s := New()

	err := s.Begin(model.FuelJetA1, model.FuelJetA, neverConfirmed, 50*time.Millisecond)
	if err != model.ErrSwitchTimeout {
		t.Fatalf("expected ErrSwitchTimeout, got %v", err)
	}

	// The core of the bug: after the timeout the switch must be back in idle
	// so it can be retried, not stuck in draining.
	if got := s.State(); got != model.SwitchIdle {
		t.Fatalf("expected state idle after timeout, got %q", got)
	}
	if s.InProgress() {
		t.Fatalf("expected switch not in progress after timeout")
	}
}

func TestBeginTimeoutAllowsRetry(t *testing.T) {
	s := New()

	if err := s.Begin(model.FuelJetA1, model.FuelJetA, neverConfirmed, 30*time.Millisecond); err != model.ErrSwitchTimeout {
		t.Fatalf("first begin: expected ErrSwitchTimeout, got %v", err)
	}

	// A second attempt must be accepted (previously rejected with "fuel
	// switch already in progress" because draining was never cleared).
	// Use an immediately-confirmed drain so the retry succeeds.
	if err := s.Begin(model.FuelJetA1, model.FuelJetA, func() bool { return true }, 50*time.Millisecond); err != nil {
		t.Fatalf("retry begin failed: %v", err)
	}
	if got := s.State(); got != model.SwitchSwitching {
		t.Fatalf("expected state switching after confirmed retry, got %q", got)
	}
}

func TestBeginConfirmedTransitionsToSwitching(t *testing.T) {
	s := New()

	if err := s.Begin(model.FuelJetA1, model.FuelJetA, func() bool { return true }, 50*time.Millisecond); err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if got := s.State(); got != model.SwitchSwitching {
		t.Fatalf("expected state switching, got %q", got)
	}
}
