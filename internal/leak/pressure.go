package leak

import "sort"

func (d *Detector) DetectPressureDrops(previous, current map[string]float64, threshold float64) []string {
	ids := make([]string, 0)
	for id, now := range current {
		before, known := previous[id]
		if !known {
			continue
		}
		if d.PressureDrop(now, before, threshold) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
