package routine

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

type Pool struct {
	waitGroup sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	option
}

// NewPool creates a Pool.
func NewPool(parentCtx context.Context, opts ...func(*option)) *Pool {
	ctx, cancel := context.WithCancel(parentCtx)
	p := &Pool{
		ctx:    ctx,
		cancel: cancel,
		option: option{recoverFunc: defaultRecoverGoroutine},
	}
	for _, opt := range opts {
		opt(&p.option)
	}
	return p
}

// Go starts a recoverable goroutine with a context.
func (p *Pool) Go(goroutine func(context.Context)) {
	p.waitGroup.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if p.recoverFunc != nil {
					p.recoverFunc(p.ctx, r)
				}
			}
			p.waitGroup.Done()
		}()
		goroutine(p.ctx)
	}()
}

// Wait waits all started routines, waiting for their termination.
func (p *Pool) Wait() {
	p.waitGroup.Wait()
	p.cancel()
}

// Stop stops all started routines, waiting for their termination.
func (p *Pool) Stop() {
	p.cancel()
	p.waitGroup.Wait()
}

func defaultRecoverGoroutine(_ context.Context, err interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "Error:%v\nStack: %s", err, debug.Stack())
}
