package alarm

import (
	"errors"
	"testing"
)

type verifyFailingRestore struct{}

func (verifyFailingRestore) Release(id string) error { return errors.New("writeback failed") }

func TestRestoreWritebackErrorNotSwallowed(t *testing.T) {
	m := NewManager(nil, verifyFailingRestore{})
	m.Raise("H1", "hydrant-leak")
	if err := m.Clear("H1"); err == nil {
		t.Fatal("expected restore writeback error to be reported")
	}
}
