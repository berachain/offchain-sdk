package types

// Marshallable is an interface that defines the Marshal and Unmarshal methods.
type Marshallable interface {
	New() Marshallable
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Queue[T Marshallable] interface {
	Push(T) (string, error)
	Receive() (string, T, bool)
	ReceiveMany(num int32) ([]string, []T, error)
	Delete(string) error
	Len() int
}
