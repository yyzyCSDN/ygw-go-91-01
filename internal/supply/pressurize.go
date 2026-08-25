package supply

import (
	"context"
	"time"

	"fuelhydrant/internal/model"
)

func (c *Coordinator) waitPressureStable(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	stableReadings := 0
	for {
		select {
		case <-ctx.Done():
			return model.ErrPressureUnstable
		case <-ticker.C:
			if c.stableNow() {
				stableReadings++
				if stableReadings >= 3 {
					return nil
				}
			} else {
				stableReadings = 0
			}
		}
	}
}

func (c *Coordinator) stableNow() bool {
	return c.group.PressureStable(c.target, c.tolerance)
}

func (c *Coordinator) WaitPressureStable(timeout time.Duration) error {
	return c.waitPressureStable(timeout)
}
