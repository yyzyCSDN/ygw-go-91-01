package pump

import (
	"testing"

	"fuelhydrant/internal/model"
)

func TestManualStateSyncedToAuto(t *testing.T) {
	g := NewGroup()
	p := NewPump("p1", NewMotor(), nil)
	g.AddPump(p)
	if err := g.ManualStart("p1"); err != nil {
		t.Fatal(err)
	}
	if g.Mode() != model.PumpModeManual {
		t.Fatalf("expected manual mode, got %s", g.Mode())
	}
}
