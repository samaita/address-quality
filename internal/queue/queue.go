// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

// Package queue provides a channel-backed, single-queue worker pool. Jobs are
// submitted to one buffered channel and drained by N worker goroutines. All
// queued jobs are processed before Shutdown returns.
package queue

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrQueueClosed = errors.New("queue closed")
	ErrQueueFull   = errors.New("queue full")
)

type Queue[T any] struct {
	jobs    chan T
	handler func(ctx context.Context, job T)

	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func New[T any](size, workers int, handler func(ctx context.Context, job T)) *Queue[T] {
	if size < 1 {
		size = 1
	}
	if workers < 1 {
		workers = 1
	}

	q := &Queue[T]{
		jobs:    make(chan T, size),
		handler: handler,
	}

	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.run()
	}

	return q
}

func (q *Queue[T]) run() {
	defer q.wg.Done()
	for job := range q.jobs {
		q.execute(job)
	}
}

func (q *Queue[T]) execute(job T) {
	defer func() { recover() }()
	q.handler(context.Background(), job)
}

// Submit enqueues a job without blocking. It returns ErrQueueClosed once
// Shutdown has started and ErrQueueFull when the buffer is at capacity.
func (q *Queue[T]) Submit(job T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrQueueClosed
	}

	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue[T]) Len() int {
	return len(q.jobs)
}

// Shutdown stops accepting new jobs and waits for all buffered jobs to be
// processed by the workers. It returns early with ctx.Err() if the context is
// cancelled before the queue drains.
func (q *Queue[T]) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	if !q.closed {
		q.closed = true
		close(q.jobs)
	}
	q.mu.Unlock()

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
