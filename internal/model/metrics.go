package model

func Sum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}

func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return Sum(values) / float64(len(values))
}

func AbsDelta(a, b float64) float64 {
	delta := a - b
	if delta < 0 {
		return -delta
	}
	return delta
}

func WithinRange(value, low, high float64) bool {
	return value >= low && value <= high
}

func Clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
