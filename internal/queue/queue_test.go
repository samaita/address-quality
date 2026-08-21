// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestSubmitProcessesAllJobs(t *testing.T) {
	var processed atomic.Int64
	const n = 1000
	q := New[int](n, 1, func(_ context.Context, job int) {
		processed.Add(1)
	})
	defer q.Shutdown(context.Background())

	for i := 0; i < n; i++ {
		if err := q.Submit(i); err != nil {
			t.Fatalf("Submit returned error: %v", err)
		}
	}

	waitFor(t, func() bool { return processed.Load() == n })
	if got := processed.Load(); got != n {
		t.Fatalf("processed = %d, want %d", got, n)
	}
}

func TestShutdownDrainsBufferedJobs(t *testing.T) {
	var processed atomic.Int64
	q := New[int](100, 1, func(_ context.Context, job int) {
		processed.Add(1)
	})

	const n = 100
	for i := 0; i < n; i++ {
		if err := q.Submit(i); err != nil {
			t.Fatalf("Submit returned error: %v", err)
		}
	}

	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if got := processed.Load(); got != n {
		t.Fatalf("processed = %d after Shutdown, want %d", got, n)
	}
}

func TestShutdownRespectsContext(t *testing.T) {
	q := New[int](10, 1, func(_ context.Context, job int) {
		time.Sleep(50 * time.Millisecond)
	})
	for i := 0; i < 10; i++ {
		q.Submit(i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := q.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown error = %v, want DeadlineExceeded", err)
	}
	q.Shutdown(context.Background())
}

func TestSubmitAfterShutdown(t *testing.T) {
	q := New[int](10, 1, func(_ context.Context, job int) {})
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if err := q.Submit(1); !errors.Is(err, ErrQueueClosed) {
		t.Fatalf("Submit after Shutdown error = %v, want ErrQueueClosed", err)
	}
}

func TestSubmitWhenFull(t *testing.T) {
	q := New[int](1, 1, func(_ context.Context, job int) {
		time.Sleep(10 * time.Millisecond)
	})
	defer q.Shutdown(context.Background())

	if err := q.Submit(1); err != nil {
		t.Fatalf("first Submit returned error: %v", err)
	}
	if err := q.Submit(2); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("Submit to full queue error = %v, want ErrQueueFull", err)
	}
}

func TestMultipleWorkersProcessAllJobs(t *testing.T) {
	var processed atomic.Int64
	const n = 2000
	q := New[int](n, 4, func(_ context.Context, job int) {
		processed.Add(1)
	})

	for i := 0; i < n; i++ {
		if err := q.Submit(i); err != nil {
			t.Fatalf("Submit returned error: %v", err)
		}
	}

	waitFor(t, func() bool { return processed.Load() == n })
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}

func TestWorkerPanicDoesNotStopQueue(t *testing.T) {
	var processed atomic.Int64
	q := New[int](10, 1, func(_ context.Context, job int) {
		if job == 0 {
			panic("boom")
		}
		processed.Add(1)
	})
	defer q.Shutdown(context.Background())

	for i := 0; i < 5; i++ {
		q.Submit(i)
	}

	waitFor(t, func() bool { return processed.Load() == 4 })
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met within deadline")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
