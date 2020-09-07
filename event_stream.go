package hblade

type Event struct {
	Name string
	Data interface{}
}

type EventStream struct {
	Events chan *Event
	Closed chan struct{}
}

func NewEventStream() *EventStream {
	return &EventStream{
		Events: make(chan *Event),
		Closed: make(chan struct{}),
	}
}
