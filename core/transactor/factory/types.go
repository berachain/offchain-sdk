package factory

// Noncer is an interface for acquiring fresh nonces.
type Noncer interface {
	Acquire() (uint64, bool)
}
