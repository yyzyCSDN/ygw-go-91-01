package supply

import (
	"testing"
	"time"

	"fuelhydrant/internal/hydrant"
	"fuelhydrant/internal/model"
	"fuelhydrant/internal/pump"
	"fuelhydrant/internal/switch"
)

func newTestCoordinator() *Coordinator {
	group := pump.NewGroup()
	for _, id := range []string{"pump-1", "pump-2"} {
		group.AddPump(pump.NewPump(id, pump.NewMotor(), nil))
	}
	monitor := hydrant.NewMonitor()
	switcher := fuelswitch.New()
	return NewCoordinator(monitor, group, switcher)
}

// Happy path: when the drain confirms, SwitchFuel dispatches the new fuel,
// completes the switch, and records a flush cycle.
func TestSwitchFuelConfirmedDispatchesAndFlushes(t *testing.T) {
	c := newTestCoordinator()
	// Drain is already complete, so drainDone() returns true immediately and
	// Begin's 15s timeout never fires.
	c.SetDrainProgress(1.0)

	if err := c.SwitchFuel(model.FuelJetA); err != nil {
		t.Fatalf("switch fuel failed: %v", err)
	}
	if got := c.Fuel(); got != model.FuelJetA {
		t.Fatalf("expected fuel dispatched, got %q", got)
	}
	if c.FlushCycles() != 1 {
		t.Fatalf("expected one flush cycle, got %d", c.FlushCycles())
	}
	if c.switcher.State() != model.SwitchIdle {
		t.Fatalf("expected switcher idle after complete, got %q", c.switcher.State())
	}
}

// The bug: a drain timeout must not leave the switcher stuck in draining.
// SwitchFuel's 15s timeout is impractical to drive here, so exercise the
// state machine directly — the behaviour SwitchFuel now relies on.
func TestSwitcherTimeoutResetsAndRetries(t *testing.T) {
	s := fuelswitch.New()

	err := s.Begin(model.FuelJetA1, model.FuelJetA, func() bool { return false }, 40*time.Millisecond)
	if err != model.ErrSwitchTimeout {
		t.Fatalf("expected ErrSwitchTimeout, got %v", err)
	}
	if got := s.State(); got != model.SwitchIdle {
		t.Fatalf("expected idle after timeout, got %q", got)
	}
	if s.InProgress() {
		t.Fatalf("expected switch not in progress after timeout")
	}
	// Retry must be accepted — previously rejected because draining was never
	// cleared.
	if err := s.Begin(model.FuelJetA1, model.FuelJetA, func() bool { return true }, 40*time.Millisecond); err != nil {
		t.Fatalf("retry begin failed: %v", err)
	}
	if got := s.State(); got != model.SwitchSwitching {
		t.Fatalf("expected switching after confirmed retry, got %q", got)
	}
}
