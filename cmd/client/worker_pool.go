package main

import (
	"context"
	"log"
	"sync"

	desc "PWZ1.0/pkg/pwz"
)

type processJob struct {
	orderIDs []uint64
}

type WorkerPool struct {
	mu          sync.Mutex
	workers     int
	jobChan     chan processJob
	errChan     chan error
	cancelFuncs []context.CancelFunc

	client desc.NotifierClient
	userID uint64
	action desc.ActionType
}

func NewWorkerPool(client desc.NotifierClient, userID uint64, action desc.ActionType) *WorkerPool {
	return &WorkerPool{
		jobChan: make(chan processJob, 100),
		errChan: make(chan error, 100),
		client:  client,
		userID:  userID,
		action:  action,
	}
}

func (wp *WorkerPool) Start(n int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for i := 0; i < n; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		wp.cancelFuncs = append(wp.cancelFuncs, cancel)
		go wp.worker(ctx, len(wp.cancelFuncs))
	}
	wp.workers += n
	log.Printf("[WorkerPool] started %d new workers (total %d)", n, wp.workers)
}

func (wp *WorkerPool) Stop(n int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if n > wp.workers {
		n = wp.workers
	}

	for i := 0; i < n; i++ {
		cancel := wp.cancelFuncs[0]
		cancel()
		wp.cancelFuncs = wp.cancelFuncs[1:]
		wp.workers--
	}
	log.Printf("[WorkerPool] stopped %d workers (total %d)", n, wp.workers)
}

func (wp *WorkerPool) Update(newCount int) {
	wp.mu.Lock()
	current := wp.workers
	wp.mu.Unlock()

	diff := newCount - current
	if diff > 0 {
		wp.Start(diff)
	} else if diff < 0 {
		wp.Stop(-diff)
	} else {
		log.Println("[WorkerPool] worker count unchanged")
	}
}

func (wp *WorkerPool) Submit(job processJob) {
	wp.jobChan <- job
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	log.Printf("[Worker %d] started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] stopped", id)
			return
		case job, ok := <-wp.jobChan:
			if !ok {
				log.Printf("[Worker %d] channel closed", id)
				return
			}

			err := processOrders(context.Background(), wp.client, wp.userID, wp.action, job.orderIDs)
			if err != nil {
				log.Printf("[Worker %d] error: %v", id, err)
				wp.errChan <- err
			} else {
				log.Printf("[Worker %d] processed orders: %v", id, job.orderIDs)
			}
		}
	}
}

func (wp *WorkerPool) Shutdown() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	close(wp.jobChan)

	for _, cancel := range wp.cancelFuncs {
		cancel()
	}
	wp.workers = 0

	log.Println("[WorkerPool] shutdown complete")
}
