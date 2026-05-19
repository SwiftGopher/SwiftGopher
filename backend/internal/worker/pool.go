package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

type Job func()

type Pool struct {
	jobs chan Job
	size int
	log  *slog.Logger

	wg sync.WaitGroup

	mu     sync.RWMutex
	closed bool

	startOnce sync.Once

	workersWg sync.WaitGroup
}

func NewPool(size int, queueSize int, log *slog.Logger) *Pool {
	return &Pool{
		jobs: make(chan Job, queueSize),
		size: size,
		log:  log,
	}
}
func (p *Pool) Start(ctx context.Context) {
	p.startOnce.Do(func() {
		p.workersWg.Add(p.size)

		for i := 0; i < p.size; i++ {
			go func(id int) {
				defer p.workersWg.Done()
				p.worker(ctx, id)
			}(i)
		}
	})
}
func (p *Pool) Submit(job Job) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return errors.New("worker pool is closed")
	}
	p.mu.RUnlock()

	p.wg.Add(1)
	select {
	case p.jobs <- job:
		return nil
	default:
		p.wg.Done() // undo the Add since we're not submitting
		return errors.New("queue full")
	}
}
func (p *Pool) worker(ctx context.Context, id int) {
	p.log.Debug("worker started", "worker_id", id)

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				p.log.Debug("worker stopped", "worker_id", id)
				return
			}
			p.run(job, id)

		case <-ctx.Done():
			p.log.Debug("worker stopped by context", "worker_id", id)
			return
		}
	}
}
func (p *Pool) run(job Job, workerID int) {
	defer p.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			p.log.Error("panic recovered",
				"worker_id", workerID,
				"panic", r,
			)
		}
	}()

	job()
}
func (p *Pool) Shutdown(ctx context.Context) {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.jobs)
	}
	p.mu.Unlock()

	done := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.log.Info("worker pool shutdown complete")
	case <-ctx.Done():
		p.log.Warn("worker pool shutdown timeout")
	}
}
func (p *Pool) Wait() {
	p.wg.Wait()
}
