package worker

import (
	"log"
	"sync"
)

// Task is a function that performs the job and returns an error.
type Task func() error

type WorkerPool struct {
	taskQueue chan Task
	wg        sync.WaitGroup
}

// NewWorkerPool creates a new worker pool with a given number of workers.
func NewWorkerPool(numWorkers int, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		taskQueue: make(chan Task, queueSize),
	}

	// Start the workers
	for i := 0; i < numWorkers; i++ {
		go pool.worker(i)
	}

	return pool
}

// AddTask adds a new task to the task queue.
func (p *WorkerPool) AddTask(task Task) error {
	result := make(chan error, 1)

	// Wrap the task with a result channel
	taskWithResult := func() error {
		err := task()
		result <- err
		close(result)
		return err
	}

	p.wg.Add(1)
	p.taskQueue <- taskWithResult

	// Wait for the task to complete and get the result.
	err := <-result
	p.wg.Done()

	return err
}

// worker listens for tasks and processes them.
func (p *WorkerPool) worker(id int) {
	for task := range p.taskQueue {
		log.Printf("Worker %d: Starting task", id)
		task()
		log.Printf("Worker %d: Task completed", id)
	}
}

// Stop gracefully stops the worker pool.
func (p *WorkerPool) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
}
