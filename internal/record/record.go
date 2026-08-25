package record

import (
	"bufio"
	"errors"
	"os"
	"sync"
)

var ErrNotOpen = errors.New("journal not open")

type Journal struct {
	mu          sync.Mutex
	path        string
	file        *os.File
	writer      *bufio.Writer
	closed      bool
	openHandles int
}

func NewJournal(path string) *Journal {
	return &Journal{path: path}
}

func (j *Journal) Path() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.path
}

func (j *Journal) Open() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.openLocked()
}

func (j *Journal) openLocked() error {
	if j.file != nil {
		return nil
	}
	f, err := os.OpenFile(j.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	j.file = f
	j.writer = bufio.NewWriter(f)
	j.closed = false
	j.openHandles++
	return nil
}

func (j *Journal) Append(kind, body string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.writer == nil {
		return ErrNotOpen
	}
	_, err := j.writer.WriteString(kind + "\t" + body + "\n")
	return err
}

func (j *Journal) Flush() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.writer == nil {
		return ErrNotOpen
	}
	return j.writer.Flush()
}

func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.closeLocked()
}

func (j *Journal) closeLocked() error {
	if j.file == nil {
		return nil
	}
	j.closed = true
	return j.flushAndCloseLocked()
}

func (j *Journal) flushAndCloseLocked() error {
	flushErr := j.writer.Flush()
	closeErr := j.file.Close()
	j.file = nil
	j.writer = nil
	j.openHandles--
	if flushErr != nil {
		return flushErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func (j *Journal) OpenHandles() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.openHandles
}

func (j *Journal) Closed() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.closed
}
