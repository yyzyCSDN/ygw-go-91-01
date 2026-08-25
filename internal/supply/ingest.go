package supply

import (
	"fuelhydrant/internal/model"
)

func (c *Coordinator) IngestTelemetry(t model.Telemetry) error {
	if !model.ValidPressure(t.Pressure) {
		return model.ErrPressureUnstable
	}
	c.monitor.ApplyTelemetry(t)
	c.group.SyncHydrantPressure(t.Hydrant, t.Pressure)
	return nil
}

func (c *Coordinator) AveragePressure() float64 {
	return c.monitor.AveragePressure()
}

func (c *Coordinator) ConsumptionRate() float64 {
	return c.monitor.TotalFlow()
}

func (c *Coordinator) PressureTarget() (float64, float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.target, c.tolerance
}

func (c *Coordinator) ApplyTelemetryBatch(items []model.Telemetry) {
	for _, item := range items {
		_ = c.IngestTelemetry(item)
	}
}
