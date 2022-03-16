package async

import (
	"context"
	"fmt"
)

type ManagerCallback interface {
	Register(name string, callback Callback)
	Unregister(name string)
	Run(ctx context.Context, param *Param) error
}

// Callback your business should implement it
type Callback interface {
	Run(ctx context.Context, param *Param) error
}

func NewManager() ManagerCallback {
	return &manager{callbacks: make(map[string]Callback, 40)}
}

type manager struct {
	callbacks map[string]Callback
}

// Register unsafe
func (m *manager) Register(name string, callback Callback) {
	m.callbacks[name] = callback
}

// Unregister unsafe
func (m *manager) Unregister(name string) {
	delete(m.callbacks, name)
}

func (m *manager) Run(ctx context.Context, param *Param) error {
	callback, found := m.callbacks[param.TaskType]
	if !found {
		return fmt.Errorf("must register %s", param.TaskType)
	}
	function, ok := callback.(Callback)
	if !ok {
		return fmt.Errorf("%s must impl Callback interface", param.TaskType)
	}
	return function.Run(ctx, param)
}
