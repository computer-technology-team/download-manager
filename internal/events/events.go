package events

type EventType int

const (
	chanSize = 100

	DownloadFailed EventType = iota
	DownloadProgressed
	DownloadCompleted
	DownloadStateChanged

	QueueCreated
	QueueDeleted
	QueueEdited
	DownloadCreated
	DownloadDeleted
)

type Event struct {
	EventType EventType
	Payload   interface{}
}

var eventChannel, uiEventChannel chan Event

func GetEventChannel() chan Event {
	if eventChannel == nil {
		eventChannel = make(chan Event, chanSize)
	}
	return eventChannel
}

// UI socket
func GetUIEventChannel() chan Event {
	if uiEventChannel == nil {
		uiEventChannel = make(chan Event, chanSize)
	}
	return uiEventChannel
}
