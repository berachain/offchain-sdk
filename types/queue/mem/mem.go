package mem

import (
	"container/list"
	"sync"

	"github.com/berachain/go-utils/utils"
	"github.com/berachain/offchain-sdk/v2/types/queue/types"
)

// Queue is a thread-safe FIFO queue implementation.
type Queue[T types.Marshallable] struct {
	mu          sync.RWMutex
	queuedItems *list.List
}

// NewQueue creates a new Queue instance.
func NewQueue[T types.Marshallable]() *Queue[T] {
	return &Queue[T]{queuedItems: list.New()}
}

// Push adds a value to the back of the queue.
func (q *Queue[T]) Push(val T) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queuedItems.PushBack(val)
	return val.String(), nil
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
	val := utils.MustGetAs[T](element.Value)
	return val.String(), val, true
}

func (q *Queue[T]) ReceiveMany(num int32) ([]string, []T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var (
		msgIDs     []string
		txRequests []T
	)
	for i := int32(0); i < num; i++ {
		element := q.queuedItems.Front()
		if element == nil {
			break
		}
		q.queuedItems.Remove(element)
		val := utils.MustGetAs[T](element.Value)
		msgIDs = append(msgIDs, val.String())
		txRequests = append(txRequests, val)
	}
	return msgIDs, txRequests, nil
}

// Delete is no-op for the in-memory queue (already deleted by receiving).
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
