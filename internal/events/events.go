package events

type EventType int

const(
	chanSize = 100

	DownloadFailed EventType = iota
	DownloadProgressed

	QueueCreated
	QueueDeleted
	QueueEdited
	DownloadCreated
	DownloadDeleted
)

type Event struct {
	EventType EventType
	Payload interface{}
}

var eventChannel chan Event

func GetChannel() chan Event {
	if eventChannel == nil {
		eventChannel = make(chan Event, chanSize)
	}
	return eventChannel
}