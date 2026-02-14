package worker

import (
	"context"
	"log/slog"
	"sync"
)

type Task func()

type Pool struct {
	maxWorkers int
	taskQueue  chan Task
	wg         sync.WaitGroup
	log        *slog.Logger
}

func NewPool(maxWorkers int, queueSize int, log *slog.Logger) *Pool {
	return &Pool{
		maxWorkers: maxWorkers,
		taskQueue:  make(chan Task, queueSize),
		log:        log.With(slog.String("component", "worker_pool")),
	}
}

func (p *Pool) Start(ctx context.Context) {
	p.log.Info("starting worker pool", "workers", p.maxWorkers)
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			p.log.Debug("worker shutting down", "worker_id", id)
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			task()
		}
	}
}

func (p *Pool) Submit(t Task) {
	p.taskQueue <- t
}

func (p *Pool) Shutdown() {
	close(p.taskQueue)
	p.wg.Wait()
	p.log.Info("worker pool shut down cleanly")
}
