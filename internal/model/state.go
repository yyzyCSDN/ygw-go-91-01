package model

func (s SupplyState) IsIdle() bool {
	return s == SupplyIdle
}

func (s SupplyState) IsSupplying() bool {
	return s == SupplySupplying
}

func ValidPressure(value float64) bool {
	return value >= 0
}
