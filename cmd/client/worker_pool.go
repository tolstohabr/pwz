package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	desc "PWZ1.0/pkg/pwz"
)

type processJob struct {
	orderIDs []uint64
}

func processOrdersAsync(
	ctx context.Context,
	client desc.NotifierClient,
	userID uint64,
	action desc.ActionType,
	orderIDs []uint64,
	batchSize int,
	workers int,
) error {

	var jobs []processJob
	for i := 0; i < len(orderIDs); i += batchSize {
		end := i + batchSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		jobs = append(jobs, processJob{
			orderIDs: orderIDs[i:end],
		})
	}

	jobChan := make(chan processJob)
	errChan := make(chan error, len(jobs))
	wg := sync.WaitGroup{}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobChan {
				err := processOrders(ctx, client, userID, action, job.orderIDs)
				if err != nil {
					log.Printf("[Worker %d] error: %v", workerID, err)
					errChan <- err
				} else {
					log.Printf("[Worker %d] processed orders: %v", workerID, job.orderIDs)
				}
			}
		}(w)
	}

	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	wg.Wait()
	close(errChan)

	var hasError bool
	for err := range errChan {
		if err != nil {
			hasError = true
			log.Println("ошибка в job:", err)
		}
	}

	if hasError {
		return fmt.Errorf("ошибки в jobах")
	}

	return nil
}
