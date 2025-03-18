package queues

// this is the abstractions exposed to UI
type QueueManager struct {
}

// This should be a singleton class
// constructor
func NewQueueManager(queue_manager *QueueManager) *QueueManager {
	return &QueueManager{}
}

func init() {
	// intialize everything upon startup
}

// use to initialise a queue
func CreateQueue() bool {

	return true
}

// use to delete a queue
func DeleteQueue() bool {
	return true
}

// use to get a single queue
func GetQueue() {

}

// list all queues
func ListQueues() {

}
