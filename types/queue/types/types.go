package types

import (
	"fmt"
	"time"
)

// Marshallable is an interface that defines the Marshal and Unmarshal methods.
type Marshallable interface {
	fmt.Stringer
	New() Marshallable
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Queue[T Marshallable] interface {
	InQueue(messageID string) bool
	Push(T) (string, error)
	Receive() (string, T, time.Time, bool)
	ReceiveMany(num int32) ([]string, []T, []time.Time, error)
	Delete(string) error
	Len() int
}
