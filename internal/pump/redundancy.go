package pump

import "fuelhydrant/internal/model"

func (g *Group) Failover(primary, backup string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	primaryPump, ok := g.pumps[primary]
	if !ok {
		return model.ErrHydrantNotFound
	}
	backupPump, ok := g.pumps[backup]
	if !ok {
		return model.ErrHydrantNotFound
	}
	if primaryPump.State() == model.PumpRunning {
		return nil
	}
	if err := backupPump.Start(); err != nil {
		return err
	}
	g.running[backup] = true
	g.running[primary] = false
	return nil
}

func (g *Group) Health() map[string]model.PumpState {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make(map[string]model.PumpState)
	for id, p := range g.pumps {
		out[id] = p.State()
	}
	return out
}

func (g *Group) RunningCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	count := 0
	for _, p := range g.pumps {
		if p.State() == model.PumpRunning {
			count++
		}
	}
	return count
}

func (g *Group) Summary() map[string]any {
	g.mu.Lock()
	defer g.mu.Unlock()
	running := make([]string, 0)
	for id, p := range g.pumps {
		if p.State() == model.PumpRunning {
			running = append(running, id)
		}
	}
	return map[string]any{
		"mode":     g.mode,
		"fuel":     g.fuel,
		"pressure": g.pressure,
		"running":  running,
	}
}
