package event

// Dispatcher is a generic event dispatcher. It maintains a mapping of unique indexes to
// subscribers, which are channels that events are sent to.
type Dispatcher[E any] struct {
	subscribers map[int]chan E
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher[E any]() *Dispatcher[E] {
	return &Dispatcher[E]{
		subscribers: make(map[int]chan E),
	}
}

// Subscribe adds a new subscriber to the Dispatcher. The subscriber is a channel on which events
// will be sent to. Returns the unique index of the subscriber.
func (d *Dispatcher[E]) Subscribe(subscriber chan E) int {
	index := len(d.subscribers)
	d.subscribers[index] = subscriber
	return index
}

// Unsubscribe removes a subscriber from the Dispatcher at the given unique index.
func (d *Dispatcher[E]) Unsubscribe(index int) {
	delete(d.subscribers, index)
}

// Dispatch sends an event to all subscribers.
func (d *Dispatcher[E]) Dispatch(event E) {
	for _, subscriber := range d.subscribers {
		subscriber <- event
	}
}
