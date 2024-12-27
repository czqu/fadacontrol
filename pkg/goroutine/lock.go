package goroutine

import (
	"context"
	"sync"
	"time"
)

type Blocker struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	ctx    context.Context
}

func NewBlocker(timeout time.Duration) *Blocker {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &Blocker{
		cancel: cancel,
		ctx:    ctx,
	}
}

func (b *Blocker) Wait() {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-b.ctx.Done():
		return
	}
}

func (b *Blocker) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Blocker) Reset(timeout time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	b.ctx = ctx
	b.cancel = cancel
}
