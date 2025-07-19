package chassis

import (
	"context"
	"sync"
	"time"
)

var shutdownWaiter sync.WaitGroup

// RegisterShutdownFunc adds a function which should be called after the passed
// context has been canceled. This function should be called in conjunction with
// WaitUntilCleanShutdown to ensure all passed shutdown functions have had time
// to exit cleanly.
func RegisterShutdownFunc(ctx context.Context, f func()) (stop func() bool) {
	shutdownWaiter.Add(1)

	return context.AfterFunc(ctx, func() {
		f()
		shutdownWaiter.Done()
	})
}

// WaitUntilCleanShutdown waits until all shutdown functions have completed
// or the supplied duration passes, which ever happens first.
func WaitUntilCleanShutdown(timeout time.Duration) {
	terminate := make(chan struct{})
	ticker := time.NewTicker(timeout)

	go func() {
		shutdownWaiter.Wait()
		terminate <- struct{}{}
	}()

	select {
	case <-terminate:
		break
	case <-ticker.C:
		break
	}

	ticker.Stop()
}
