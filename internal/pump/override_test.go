package pump

import (
	"errors"
	"testing"

	"fuelhydrant/internal/model"
)

func TestManualStartSyncsModeAndYieldsAuto(t *testing.T) {
	g := NewGroup()
	g.AddPump(NewPump("pump-1", NewMotor(), nil))
	g.AddPump(NewPump("pump-2", NewMotor(), nil))

	if g.IsManual() {
		t.Fatalf("fresh group should be in auto mode")
	}

	// Operator hand-starts pump-1 on site.
	if err := g.ManualStart("pump-1"); err != nil {
		t.Fatalf("manual start: %v", err)
	}

	// State must sync to manual so auto logic can see the override.
	if !g.IsManual() {
		t.Fatalf("after manual start group mode=%s, want manual", g.Mode())
	}
	if g.State("pump-1") != model.PumpRunning {
		t.Fatalf("pump-1 state=%s, want running", g.State("pump-1"))
	}

	// Auto StartAll must yield, not overwrite or restart the manual pump.
	if err := g.StartAll(); err != nil {
		t.Fatalf("StartAll during manual: %v", err)
	}
	if g.State("pump-1") != model.PumpRunning {
		t.Fatalf("StartAll overwrote manual pump-1: %s", g.State("pump-1"))
	}
	if g.State("pump-2") == model.PumpRunning {
		t.Fatalf("StartAll started pump-2 during manual override")
	}

	// Auto StopAll must NOT stop the manually controlled pump.
	if err := g.StopAll(); err != nil {
		t.Fatalf("StopAll during manual: %v", err)
	}
	if g.State("pump-1") != model.PumpRunning {
		t.Fatalf("auto StopAll stopped manual pump-1: %s", g.State("pump-1"))
	}

	// ApplyFuelSequence must not drop the manual pump's running state.
	g.ApplyFuelSequence(model.FuelAvgas)
	if g.State("pump-1") != model.PumpRunning {
		t.Fatalf("ApplyFuelSequence dropped manual pump-1: %s", g.State("pump-1"))
	}
}

func TestForcedStopOverridesManualForSafety(t *testing.T) {
	g := NewGroup()
	g.AddPump(NewPump("pump-1", NewMotor(), nil))

	_ = g.ManualStart("pump-1")
	if g.State("pump-1") != model.PumpRunning {
		t.Fatalf("manual start failed")
	}

	// Safety interlock path must halt every pump including manual ones.
	if err := g.StopAllForced(); err != nil {
		t.Fatalf("StopAllForced: %v", err)
	}
	if g.State("pump-1") != model.PumpIdle {
		t.Fatalf("StopAllForced left manual pump running: %s", g.State("pump-1"))
	}
	// Group returns to auto once no pump is manually held.
	if g.IsManual() {
		t.Fatalf("group still manual after forced stop cleared all overrides")
	}
}

func TestManualStopReturnsToAutoWhenNoManualRemains(t *testing.T) {
	g := NewGroup()
	g.AddPump(NewPump("pump-1", NewMotor(), nil))
	g.AddPump(NewPump("pump-2", NewMotor(), nil))

	_ = g.ManualStart("pump-1")
	_ = g.ManualStart("pump-2")
	if !g.IsManual() {
		t.Fatalf("expected manual after two manual starts")
	}

	// Stopping one manual pump leaves the other under manual control.
	_ = g.ManualStop("pump-1")
	if !g.IsManual() {
		t.Fatalf("group returned to auto while pump-2 still manual")
	}

	// Stopping the last manual pump returns the group to auto.
	_ = g.ManualStop("pump-2")
	if g.IsManual() {
		t.Fatalf("group still manual after all manual pumps stopped")
	}
}

func TestErrManualOverrideIsSentinel(t *testing.T) {
	if !errors.Is(model.ErrManualOverride, model.ErrManualOverride) {
		t.Fatalf("ErrManualOverride must be matchable with errors.Is")
	}
}
