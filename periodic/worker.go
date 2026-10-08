package periodic

import (
	"sync"
	"time"
)

// Worker calls a function at a fixed interval until stopped.
type Worker struct {
	interval  time.Duration
	startOnce sync.Once      // ensures worker start only once
	stopOnce  sync.Once      // ensures worker stop only once
	done      chan struct{}  // to signal worker to stop
	wg        sync.WaitGroup // ensures Stop() to wait for the worker to finished
}

// New returns a Worker that will tick every interval.
// An interval of zero or less makes Run a no-op.
func New(interval time.Duration) *Worker {
	return &Worker{
		done:     make(chan struct{}),
		interval: interval,
	}
}

// Run starts calling fn every interval in a new goroutine.
// It has an effect only the first time it is called.
// It must not be called concurrently with Stop().
// fn must be safe to call from another goroutine.
func (w *Worker) Run(fn func()) {
	if w.interval <= 0 {
		return
	}

	w.startOnce.Do(func() {
		w.wg.Add(1)

		go func() {
			defer w.wg.Done()

			ticker := time.NewTicker(w.interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					fn()
				case <-w.done:
					return
				}
			}
		}()
	})
}

// Stop signals the worker and waits for it to exit.
// It is safe to call more than once and from multiple goroutines.
func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		close(w.done)
		w.wg.Wait()
	})
}
