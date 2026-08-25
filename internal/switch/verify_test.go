package fuelswitch

import (
	"testing"
	"time"

	"fuelhydrant/internal/model"
)

func TestSwitchTimeoutRecovers(t *testing.T) {
	s := New()
	err := s.Begin(model.FuelJetA1, model.FuelJetA, func() bool { return false }, 40*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if s.State() != model.SwitchIdle {
		t.Fatalf("expected switch to reset to idle, got %s", s.State())
	}
}
