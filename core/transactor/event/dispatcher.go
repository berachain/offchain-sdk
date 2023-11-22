package event

// Event is an interface that all events should implement.
// It requires an ID method that should return a unique identifier for the event.
type Event interface {
	any
}

// Dispatcher is a generic event dispatcher.
// It maintains a list of subscribers which are channels that events are sent to.
type Dispatcher[E any] struct {
	subscribers []chan E
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher[E any]() *Dispatcher[E] {
	return &Dispatcher[E]{
		subscribers: make([]chan E, 0),
	}
}

// Subscribe adds a new subscriber to the Dispatcher.
// The subscriber is a channel that events will be sent to.
func (d *Dispatcher[E]) Subscribe(subscriber chan E) {
	d.subscribers = append(d.subscribers, subscriber)
}

// Unsubscribe removes a subscriber from the Dispatcher.
// The subscriber is a channel that events will no longer be sent to.
// The for loop kinda hood, but in practice he length of `subscribers` is
// going to be so small it doesn't really matter.
func (d *Dispatcher[E]) Unsubscribe(subscriber chan E) {
	for i, s := range d.subscribers {
		if s == subscriber {
			d.subscribers = append(d.subscribers[:i], d.subscribers[i+1:]...)
			return
		}
	}
}

// Dispatch sends an event to all subscribers.
func (d *Dispatcher[E]) Dispatch(event E) {
	for _, subscriber := range d.subscribers {
		subscriber <- event
	}
}
