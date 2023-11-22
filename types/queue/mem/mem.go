package mem

import (
	"container/list"
	"sync"

	"github.com/google/uuid"
)

// Queue is a thread-safe FIFO queue implementation.
type Queue[T any] struct {
	mu              sync.RWMutex
	queuedItems     *list.List
	inProgressItems map[string]*list.Element
}

// NewQueue creates a new Queue instance.
func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{
		queuedItems:     list.New(),
		inProgressItems: make(map[string]*list.Element),
	}
	return q
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
	msgID := uuid.New().String()
	q.inProgressItems[msgID] = element

	return msgID, element.Value.(T), true
}

func (q *Queue[T]) ReceiveMany(num int32) ([]string, []T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var msgIDs []string
	var txRequests []T

	for i := int32(0); i < num; i++ {
		if element := q.queuedItems.Front(); element != nil {
			q.queuedItems.Remove(element)
			msgID := uuid.New().String()
			q.inProgressItems[msgID] = element
			msgIDs = append(msgIDs, msgID)
			txRequests = append(txRequests, element.Value.(T))
		}
	}

	return msgIDs, txRequests, nil
}

// MarkComplete removes the value associated with the given MessageID from the inProgress map.
func (q *Queue[T]) Delete(msgID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if element, ok := q.inProgressItems[msgID]; ok {
		delete(q.inProgressItems, msgID)
		q.queuedItems.Remove(element)
	}
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
