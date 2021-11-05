package routine

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type errGroup struct {
	ctx       context.Context
	err       error
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup
	errOnce   sync.Once
}

// NewGroup starts a recoverable goroutine errGroup with a context.
func NewGroup(ctx context.Context) *errGroup {
	newCtx, cancel := context.WithCancel(ctx)
	return &errGroup{
		ctx:    newCtx,
		cancel: cancel,
	}
}

// Go starts a recoverable goroutine with a context.
func (e *errGroup) Go(goroutine func(context.Context) error) {
	e.waitGroup.Add(1)
	go func() {
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v.Stack:%s", r, debug.Stack())
			}
			if err != nil {
				e.errOnce.Do(func() {
					e.err = err
					e.cancel()
				})
			}
			e.waitGroup.Done()
		}()
		err = goroutine(e.ctx)
	}()
}

func (e *errGroup) Wait() error {
	e.waitGroup.Wait()
	e.errOnce.Do(e.cancel)
	return e.err
}
