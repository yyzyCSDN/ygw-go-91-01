package supply

import (
	"errors"
	"testing"

	"fuelhydrant/internal/hydrant"
	"fuelhydrant/internal/model"
	"fuelhydrant/internal/pump"
	"fuelhydrant/internal/switch"
)

func newTestCoordinator(t *testing.T) (*Coordinator, *pump.Group) {
	t.Helper()
	group := pump.NewGroup()
	group.AddPump(pump.NewPump("pump-1", pump.NewMotor(), nil))
	group.AddPump(pump.NewPump("pump-2", pump.NewMotor(), nil))
	monitor := hydrant.NewMonitor()
	for _, id := range []string{"hydrant-1"} {
		monitor.Register(id, model.FuelJetA1)
	}
	switcher := fuelswitch.New()
	return NewCoordinator(monitor, group, switcher), group
}

func TestAutoSupplyYieldsDuringManualOverride(t *testing.T) {
	coord, group := newTestCoordinator(t)

	// Operator hand-starts pump-1 on site.
	if err := group.ManualStart("pump-1"); err != nil {
		t.Fatalf("manual start: %v", err)
	}
	if coord.AutoAllowed() {
		t.Fatalf("AutoAllowed should be false during manual override")
	}

	// Auto Start must yield with ErrManualOverride, not overwrite the pump.
	err := coord.Start()
	if !errors.Is(err, model.ErrManualOverride) {
		t.Fatalf("Start() during manual: err=%v, want ErrManualOverride", err)
	}
	if group.State("pump-1") != model.PumpRunning {
		t.Fatalf("auto Start overwrote manual pump-1: %s", group.State("pump-1"))
	}
	if coord.State() != model.SupplyIdle {
		t.Fatalf("supply state advanced past idle during override: %s", coord.State())
	}

	// Fuel switch must also yield, not re-dispatch over the manual pump.
	err = coord.SwitchFuel(model.FuelAvgas)
	if !errors.Is(err, model.ErrManualOverride) {
		t.Fatalf("SwitchFuel() during manual: err=%v, want ErrManualOverride", err)
	}
	if group.State("pump-1") != model.PumpRunning {
		t.Fatalf("SwitchFuel overwrote manual pump-1: %s", group.State("pump-1"))
	}
}

func TestAutoResumesAfterOperatorReturnsToAuto(t *testing.T) {
	coord, group := newTestCoordinator(t)

	_ = group.ManualStart("pump-1")

	// While manual, auto yields.
	if err := coord.Start(); !errors.Is(err, model.ErrManualOverride) {
		t.Fatalf("expected yield, got %v", err)
	}

	// Operator stops the manual pump -> group back to auto -> auto can resume.
	_ = group.ManualStop("pump-1")
	if !coord.AutoAllowed() {
		t.Fatalf("AutoAllowed should be true once operator returns to auto")
	}
}
