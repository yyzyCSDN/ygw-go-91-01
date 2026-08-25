package model

import "errors"

var (
	ErrPumpStartFailed  = errors.New("pump start failed")
	ErrHydrantNotFound  = errors.New("hydrant not found")
	ErrSwitchTimeout    = errors.New("fuel switch drain timeout")
	ErrPressureUnstable = errors.New("pressure not stable")
	ErrInterlockLocked  = errors.New("interlock locked")
	ErrRestoreFailed    = errors.New("restore writeback failed")
)
