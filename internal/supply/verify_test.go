package supply

import (
	"testing"
	"time"

	"fuelhydrant/internal/hydrant"
	"fuelhydrant/internal/pump"
	"fuelhydrant/internal/switch"
)

func TestPressWaitTimeoutHandled(t *testing.T) {
	monitor := hydrant.NewMonitor()
	group := pump.NewGroup()
	switcher := fuelswitch.New()
	c := NewCoordinator(monitor, group, switcher)
	c.SetTarget(4.0, 0.2)
	err := c.WaitPressureStable(40 * time.Millisecond)
	if err == nil {
		t.Fatal("expected pressure timeout error, got nil")
	}
}
