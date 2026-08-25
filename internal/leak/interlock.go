package leak

import (
	"sync"

	"fuelhydrant/internal/model"
)

type Interlock struct {
	mu    sync.Mutex
	state model.InterlockState
}

func NewInterlock() *Interlock {
	return &Interlock{state: model.InterlockFree}
}

func (i *Interlock) State() model.InterlockState {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.state
}

func (i *Interlock) Lock() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.state == model.InterlockLocked {
		return nil
	}
	i.state = model.InterlockLocked
	return nil
}

func (i *Interlock) Release() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.state != model.InterlockLocked {
		return model.ErrInterlockLocked
	}
	i.state = model.InterlockReleased
	return nil
}

func (i *Interlock) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.state = model.InterlockFree
}
