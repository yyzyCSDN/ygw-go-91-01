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
	j.file = nil
	j.writer = nil
	j.path = newPath
	return j.openLocked()
}
