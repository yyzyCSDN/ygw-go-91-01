package record

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestRunHandleClosed(t *testing.T) {
	dir := t.TempDir()
	j := NewJournal(filepath.Join(dir, "run.log"))
	if err := j.Open(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if err := j.Rotate(filepath.Join(dir, fmt.Sprintf("r%d.log", i))); err != nil {
			t.Fatal(err)
		}
	}
	if got := j.OpenHandles(); got != 1 {
		t.Fatalf("expected 1 open handle, got %d", got)
	}
	_ = j.Close()
}
