package record

import (
	"fmt"
	"strings"
)

func (j *Journal) Rotate(newPath string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if strings.TrimSpace(newPath) == "" {
		return fmt.Errorf("rotate path must not be empty")
	}
	if err := j.closeLocked(); err != nil {
		return fmt.Errorf("rotate: close current journal: %w", err)
	}
	j.path = newPath
	if err := j.openLocked(); err != nil {
		return fmt.Errorf("rotate: open next journal: %w", err)
	}
	return nil
}
