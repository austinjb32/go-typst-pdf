package queue

import (
	"go-typst-pdf/pdf"
	"log"
	"sync"
)

type Job struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

var jobQueue = make(chan Job, 100) // Buffered channel for job queue

func StartWorkerPool(workerCount int) {
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobQueue {
				log.Printf("Worker %d processing job: %v", workerID, job)
				_, err := pdf.GenerateAndUpload(job.Template, job.Data)
				if err != nil {
					log.Printf("Worker %d error processing job: %v", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()
	log.Println("All workers have completed their tasks.")
}

func AddJobToQueue(job Job) {
	jobQueue <- job
	log.Printf("Job added to queue: %v", job)
}

func CloseJobQueue() {
	close(jobQueue)
}
