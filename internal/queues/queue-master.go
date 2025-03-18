package queues

import (
	"fmt"
	"sync"
	"time"
)

// Constants
const (
	maxWorkers      = 3   // Maximum number of concurrent workers
	tokensPerSecond = 100 // Tokens produced per second (100 KB/s rate limit)
)

// DownloadTask represents a single download task
type DownloadTask struct {
	URL      string
	FilePath string
}

// QueueMaster manages the queue and workers
type QueueMaster struct {
	queue      []DownloadTask
	ticker     *Ticker // Use the Ticker for rate limiting
	workerDone chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex // Mutex to protect the queue
	startTime  time.Time  // Start time of the active period
	endTime    time.Time  // End time of the active period
}

// NewQueueMaster initializes a new QueueMaster
func NewQueueMaster(startTime, endTime time.Time) *QueueMaster {
	qm := &QueueMaster{
		queue:      make([]DownloadTask, 0),
		ticker:     NewTicker(), // Initialize the Ticker
		workerDone: make(chan struct{}),
		startTime:  startTime,
		endTime:    endTime,
	}
	qm.ticker.SetBandwidth(tokensPerSecond * 1024) // Set bandwidth in bytes per second (100 KB/s)
	return qm
}

// AddTask adds a new download task to the queue
func (qm *QueueMaster) AddTask(task DownloadTask) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.queue = append(qm.queue, task)
}

// Start begins processing the queue
func (qm *QueueMaster) Start() {
	// Start the Ticker
	qm.ticker.Start()

	// Start initial workers
	for i := 0; i < maxWorkers; i++ {
		qm.startNextWorker()
	}
}

// isActivePeriod checks if the current time is within the active period
func (qm *QueueMaster) isActivePeriod() bool {
	now := time.Now()
	return now.After(qm.startTime) && now.Before(qm.endTime)
}

// timeUntilNextActivePeriod calculates the duration until the next active period starts
func (qm *QueueMaster) timeUntilNextActivePeriod() time.Duration {
	now := time.Now()
	if now.Before(qm.startTime) {
		// If the current time is before the start time, wait until the start time
		return qm.startTime.Sub(now)
	} else {
		// If the current time is after the end time, wait until the start time of the next day
		nextStart := qm.startTime.Add(24 * time.Hour)
		return nextStart.Sub(now)
	}
}

// startNextWorker starts a new worker if there are tasks in the queue
func (qm *QueueMaster) startNextWorker() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if len(qm.queue) > 0 {
		// Dequeue the next task
		task := qm.queue[0]
		qm.queue = qm.queue[1:]

		// Start a new worker for the task
		qm.wg.Add(1)
		go qm.startWorker(task)
	} else {
		// No more tasks in the queue
		if len(qm.queue) == 0 && qm.wgWait() == 0 {
			close(qm.workerDone)
			qm.ticker.Quite() // Stop the Ticker when all workers are done
		}
	}
}

// startWorker starts a new worker for a download task
func (qm *QueueMaster) startWorker(task DownloadTask) {
	defer qm.wg.Done()

	// Use the DownloadFile function from download.go
	if err := DownloadFile(task.URL, task.FilePath, qm.ticker); err != nil {
		fmt.Printf("Error downloading %s: %v\n", task.URL, err)
	}

	// Start the next worker to replace this one
	qm.startNextWorker()
}

// wgWait returns the number of active workers
func (qm *QueueMaster) wgWait() int {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return len(qm.queue)
}

// Wait waits for all workers to finish
func (qm *QueueMaster) Wait() {
	<-qm.workerDone
}

func main() {
	// Define the active period (e.g., from 10:00 AM to 6:00 PM)
	startTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 10, 0, 0, 0, time.Local)
	endTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 18, 0, 0, 0, time.Local)

	// Initialize the queue master
	qm := NewQueueMaster(startTime, endTime)

	// Add download tasks to the queue
	qm.AddTask(DownloadTask{URL: "https://example.com/file1.zip", FilePath: "file1.zip"})
	qm.AddTask(DownloadTask{URL: "https://example.com/file2.zip", FilePath: "file2.zip"})
	qm.AddTask(DownloadTask{URL: "https://example.com/file3.zip", FilePath: "file3.zip"})
	qm.AddTask(DownloadTask{URL: "https://example.com/file4.zip", FilePath: "file4.zip"})
	qm.AddTask(DownloadTask{URL: "https://example.com/file5.zip", FilePath: "file5.zip"})

	// Start processing the queue
	qm.Start()

	// Wait for all downloads to complete
	qm.Wait()
	fmt.Println("All downloads completed.")
}
