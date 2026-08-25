package hydrant

import "testing"

func TestEmptyHydrantNoNilPanic(t *testing.T) {
	m := NewMonitor()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on unregistered hydrant: %v", r)
		}
	}()
	p, ok := m.PressureOf("missing")
	if ok {
		t.Fatal("expected unregistered hydrant to report not found")
	}
	if p != 0 {
		t.Fatalf("expected zero pressure, got %f", p)
	}
}
