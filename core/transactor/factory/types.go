package factory

// Noncer is an interface for acquiring nonces.
type Noncer interface {
	Acquire() (nonce uint64, isReplacing bool)
	RemoveAcquired(uint64)
}
