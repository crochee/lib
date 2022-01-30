package async

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type ParamPool interface {
	Get() *Param
	Put(*Param)
}

type ManagerExecutor interface {
	Register(executors ...Executor) error
	Run(ctx context.Context, param *Param) error
}

// Executor your business should implement it
type Executor interface {
	SafeCopy() Executor
	ID() string
	Run(ctx context.Context, data []byte) error
}

type Param struct {
	Name     string                 `json:"name" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
	Data     []byte                 `json:"data"`
}

func NewParamPool() ParamPool {
	return &defaultParamPool{pool: sync.Pool{New: func() interface{} {
		return &Param{
			Name:     "",
			Metadata: make(map[string]interface{}),
			Data:     make([]byte, 0),
		}
	}}}
}

type defaultParamPool struct {
	pool sync.Pool
}

func (d *defaultParamPool) Get() *Param {
	v, ok := d.pool.Get().(*Param)
	if !ok {
		return &Param{
			Name:     "",
			Metadata: make(map[string]interface{}),
			Data:     make([]byte, 0),
		}
	}
	return v
}

func (d *defaultParamPool) Put(param *Param) {
	param.Name = ""
	for key := range param.Metadata {
		delete(param.Metadata, key)
	}
	param.Data = param.Data[:0]
	d.pool.Put(param)
}

func NewManager() ManagerExecutor {
	return &manager{model: make(map[string]Executor)}
}

type manager struct {
	model map[string]Executor
}

func (m *manager) register(v Executor) error {
	var (
		vt   = reflect.TypeOf(v)
		name string
	)
	switch vt.Kind() {
	case reflect.Struct:
		name = vt.String()
	case reflect.Ptr:
		name = vt.Elem().String()
	default:
		return fmt.Errorf("not support %s", vt.String())
	}
	if _, ok := m.model[name]; ok {
		return fmt.Errorf("%s is enable", name)
	}
	m.model[name] = v
	return nil
}

func (m *manager) Register(executors ...Executor) error {
	for _, executor := range executors {
		if err := m.register(executor); err != nil {
			return err
		}
	}
	return nil
}

func (m *manager) Run(ctx context.Context, param *Param) error {
	v, ok := m.model[param.Name]
	if !ok {
		return fmt.Errorf("must register model %s", param.Name)
	}
	return v.SafeCopy().Run(ctx, param.Data)
}
