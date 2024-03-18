package mem

import (
	"container/list"
	"sync"
	"time"

	"github.com/berachain/go-utils/utils"
	"github.com/berachain/offchain-sdk/types/queue/types"
)

// Queue is a thread-safe FIFO queue implementation.
type Queue[T types.Marshallable] struct {
	mu            sync.RWMutex
	queuedItems   *list.List
	timesInserted map[string]time.Time
}

// NewQueue creates a new Queue instance.
func NewQueue[T types.Marshallable]() *Queue[T] {
	return &Queue[T]{
		queuedItems:   list.New(),
		timesInserted: make(map[string]time.Time),
	}
}

func (q *Queue[T]) InQueue(messageID string) bool {
	_, ok := q.timesInserted[messageID]
	return ok
}

// Push adds a value to the back of the queue.
func (q *Queue[T]) Push(val T) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queuedItems.PushBack(val)
	q.timesInserted[val.String()] = time.Now()

	return val.String(), nil
}

// Pop returns the value at the front of the queue without removing it.
// The second return value indicates if the operation succeeded.
func (q *Queue[T]) Receive() (string, T, time.Time, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	element := q.queuedItems.Front()
	if element == nil {
		return "", zeroValueOf[T](), time.Time{}, false
	}

	q.queuedItems.Remove(element)
	val := utils.MustGetAs[T](element.Value)
	msgID := val.String()
	timeInserted := q.timesInserted[msgID]
	delete(q.timesInserted, msgID)

	return msgID, val, timeInserted, true
}

func (q *Queue[T]) ReceiveMany(num int32) ([]string, []T, []time.Time, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var (
		msgIDs        []string
		txRequests    []T
		timesInserted []time.Time
	)
	for i := int32(0); i < num; i++ {
		element := q.queuedItems.Front()
		if element == nil {
			break
		}
		q.queuedItems.Remove(element)
		val := utils.MustGetAs[T](element.Value)
		msgID := val.String()
		timeInserted := q.timesInserted[msgID]
		delete(q.timesInserted, msgID)
		msgIDs = append(msgIDs, msgID)
		txRequests = append(txRequests, val)
		timesInserted = append(timesInserted, timeInserted)
	}
	return msgIDs, txRequests, timesInserted, nil
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
