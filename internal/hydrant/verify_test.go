package hydrant

import (
	"testing"

	"fuelhydrant/internal/model"
)

func TestRecoveryUsesLatestSnapshot(t *testing.T) {
	m := NewMonitor()
	m.Register("H1", model.FuelJetA1)
	m.SetLeaking("H1", true)
	m.CacheSnapshot()
	m.SetLeaking("H1", false)
	latest := m.Snapshot()
	m.Recover(latest)
	st, ok := m.Status("H1")
	if !ok {
		t.Fatal("H1 missing after recovery")
	}
	if st.Leaking {
		t.Fatal("recovered hydrant re-marked as leaking")
	}
}
