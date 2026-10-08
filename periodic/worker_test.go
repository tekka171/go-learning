package periodic

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var calls atomic.Int32
	w := New(5 * time.Millisecond)
	w.Run(func() {
		calls.Add(1)
	})

	time.Sleep(50 * time.Millisecond)
	w.Stop()

	if got := calls.Load(); got < 2 {
		t.Fatalf("fn called %d times, want at least 2", got)
	}
}

func TestRunTwice(t *testing.T) {
	var calls1 atomic.Int32
	var calls2 atomic.Int32
	w := New(5 * time.Millisecond)
	w.Run(func() {
		calls1.Add(1)
	})
	w.Run(func() {
		calls2.Add(1)
	})

	time.Sleep(50 * time.Millisecond)
	w.Stop()

	if got := calls1.Load(); got < 2 {
		t.Fatalf("fn called %d times, want at least 2", got)
	}

	if got := calls2.Load(); got > 0 {
		t.Fatalf("2nd func should not run")
	}
}

func TestRunWithZeroInterval(t *testing.T) {
	var calls atomic.Int32
	w := New(0)
	w.Run(func() {
		calls.Add(1)
	})

	time.Sleep(20 * time.Millisecond)
	w.Stop()

	if got := calls.Load(); got != 0 {
		t.Fatalf("fn called %d times, want 0", got)
	}
}

func TestStop(t *testing.T) {
	var calls atomic.Int32
	w := New(5 * time.Millisecond)
	w.Run(func() {
		calls.Add(1)
	})

	time.Sleep(50 * time.Millisecond)
	w.Stop()

	before := calls.Load()
	time.Sleep(50 * time.Millisecond)

	if after := calls.Load(); before != after {
		t.Fatalf("fn called after Stop: count went from %d to %d", before, after)
	}
}

func TestStopTwice(t *testing.T) {
	w := New(time.Second)
	w.Run(func() {})
	w.Stop()
	w.Stop()
}

func TestStopWithoutRun(t *testing.T) {
	w := New(time.Second)
	w.Stop()
}
