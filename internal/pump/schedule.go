package pump

import "sort"

type ScheduleEntry struct {
	PumpID    string
	Priority  int
	OnDemand  bool
}

func (g *Group) Schedule(entries []ScheduleEntry) []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	sorted := make([]ScheduleEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	order := make([]string, 0, len(sorted))
	for _, entry := range sorted {
		if _, ok := g.pumps[entry.PumpID]; ok {
			order = append(order, entry.PumpID)
		}
	}
	return order
}
