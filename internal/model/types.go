package model

type FuelType string

const (
	FuelJetA1 FuelType = "JET-A1"
	FuelJetA  FuelType = "JET-A"
	FuelAvgas FuelType = "AVGAS"
)

func (f FuelType) Valid() bool {
	return f == FuelJetA1 || f == FuelJetA || f == FuelAvgas
}

type SupplyState string

const (
	SupplyIdle         SupplyState = "idle"
	SupplyPressurizing SupplyState = "pressurizing"
	SupplySupplying    SupplyState = "supplying"
	SupplyStopping     SupplyState = "stopping"
)

type PumpMode string

const (
	PumpModeAuto   PumpMode = "auto"
	PumpModeManual PumpMode = "manual"
)

type PumpState string

const (
	PumpIdle    PumpState = "idle"
	PumpRunning PumpState = "running"
	PumpFailed  PumpState = "failed"
)

type InterlockState string

const (
	InterlockFree     InterlockState = "free"
	InterlockLocked   InterlockState = "locked"
	InterlockReleased InterlockState = "released"
)

type SwitchState string

const (
	SwitchIdle      SwitchState = "idle"
	SwitchDraining  SwitchState = "draining"
	SwitchSwitching SwitchState = "switching"
)

type HydrantStatus struct {
	ID       string
	Pressure float64
	Flow     float64
	Leaking  bool
	Fuel     FuelType
}

type Telemetry struct {
	Hydrant  string
	Pressure float64
	Flow     float64
}

type RecordEvent struct {
	Kind string
	Body string
}

type Snapshot struct {
	Hydrants []HydrantStatus
	Supply   SupplyState
	Fuel     FuelType
}
