package chassis

import (
	"cmp"
	"context"
	"sync"
	"testing"
	"time"

	cmpTest "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func resetWaiter() {
	shutdownWaiter = sync.WaitGroup{}
}

func Test_WaitUntilCleanShutdown(t *testing.T) {
	t.Cleanup(resetWaiter)

	t.Run("add many funcs", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		ch := make(chan int, 3)

		// Add several dummy functions
		RegisterShutdownFunc(ctx, func() { ch <- 1 })
		RegisterShutdownFunc(ctx, func() { ch <- 2 })
		RegisterShutdownFunc(ctx, func() { ch <- 3 })

		cancel()

		// Wait for all functions to terminate.
		WaitUntilCleanShutdown(func() time.Duration {
			if future, ok := t.Deadline(); ok {
				return time.Until(future)
			}

			return 5 * time.Second
		}())

		close(ch)

		values := make([]int, 0, 3)
		for v := range ch {
			values = append(values, v)
		}

		if diff := cmpTest.Diff(
			[]int{1, 2, 3}, values, cmpopts.SortSlices(cmp.Compare[int]),
		); diff != "" {
			t.Errorf("Slice mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("timeout reached", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		ch := make(chan int, 1)

		// Add a function that takes too long.
		RegisterShutdownFunc(ctx, func() {
			time.Sleep(30 * time.Second)
			ch <- 1
		})

		cancel()
		WaitUntilCleanShutdown(100 * time.Millisecond)
		close(ch)

		values := make([]int, 0, 1)
		for v := range ch {
			values = append(values, v)
		}

		if diff := cmpTest.Diff(
			[]int{}, values, cmpopts.EquateEmpty(),
		); diff != "" {
			t.Errorf("Slice mismatch (-want +got):\n%s", diff)
		}
	})
}
