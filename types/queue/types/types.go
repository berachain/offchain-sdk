package types

import (
	"fmt"
)

// Marshallable is an interface that defines the necessary methods to be (un)marshalled.
type Marshallable interface {
	fmt.Stringer
	New() Marshallable
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// Queue defines the interaction with a queue.
type Queue[T Marshallable] interface {
	Push(T) (string, error)
	Receive() (string, T, bool)
	ReceiveMany(num int32) ([]string, []T, error)
	Delete(string) error
	Len() int
}
