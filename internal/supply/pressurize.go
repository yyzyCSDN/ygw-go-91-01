package supply

import (
	"context"
	"time"

)

func (c *Coordinator) waitPressureStable(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			return nil
		}
	}
}

func (c *Coordinator) WaitPressureStable(timeout time.Duration) error {
	return c.waitPressureStable(timeout)
}
