package engine

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"sync/atomic"
)

// Coordinator manages graceful shutdown of the test engine.
type Coordinator struct {
	ctx            context.Context
	cancel         context.CancelFunc
	shutdownCh     chan struct{}
	sigCh          chan os.Signal
	softShutdown   int32 // atomic: 1 = shutdown requested
}

// NewCoordinator creates a new Coordinator.
func NewCoordinator(ctx context.Context, cancel context.CancelFunc) *Coordinator {
	return &Coordinator{
		ctx:        ctx,
		cancel:     cancel,
		shutdownCh: make(chan struct{}),
	}
}

// Start begins listening for shutdown signals.
func (c *Coordinator) Start() {
	c.sigCh = make(chan os.Signal, 1)
	signal.Notify(c.sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c.sigCh
		atomic.StoreInt32(&c.softShutdown, 1)
		close(c.shutdownCh)
	}()
}

// Stop ends signal listening and cancels the context.
func (c *Coordinator) Stop() {
	if c.sigCh != nil {
		signal.Stop(c.sigCh)
	}
	c.cancel()
}

// ResetContext replaces the active context for a new stage.
func (c *Coordinator) ResetContext(ctx context.Context, cancel context.CancelFunc) {
	c.ctx = ctx
	c.cancel = cancel
}

// ShutdownRequested returns true if a shutdown signal was received.
func (c *Coordinator) ShutdownRequested() bool {
	return atomic.LoadInt32(&c.softShutdown) == 1
}

// ShutdownChannel returns the channel that closes on shutdown.
func (c *Coordinator) ShutdownChannel() <-chan struct{} {
	return c.shutdownCh
}
