package mem

import (
	"container/list"
	"sync"
)

// Queue is a thread-safe FIFO queue implementation.
type Queue[T any] struct {
	mu          sync.RWMutex
	queuedItems *list.List
}

// NewQueue creates a new Queue instance.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		queuedItems: list.New(),
	}
}

// Push adds a value to the back of the queue.
func (q *Queue[T]) Push(val T) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queuedItems.PushBack(val)
	return "", nil
}

// Pop returns the value at the front of the queue without removing it.
// The second return value indicates if the operation succeeded.
func (q *Queue[T]) Receive() (string, T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	element := q.queuedItems.Front()
	if element == nil {
		return "", zeroValueOf[T](), false
	}

	q.queuedItems.Remove(element)
	return "", element.Value.(T), true
}

func (q *Queue[T]) ReceiveMany(num int32) ([]string, []T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var txRequests []T
	for i := int32(0); i < num; i++ {
		element := q.queuedItems.Front()
		if element == nil {
			break
		}

		q.queuedItems.Remove(element)
		txRequests = append(txRequests, element.Value.(T))
	}
	return make([]string, len(txRequests)), txRequests, nil
}

// Delete is no-op for the in-memory queue.
func (q *Queue[T]) Delete(string) error {
	return nil
}

// Len returns the number of elements currently in the queue.
func (q *Queue[T]) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.queuedItems.Len()
}

// zeroValueOf returns the zero value for a type.
func zeroValueOf[T any]() T {
	var v T
	return v
}
