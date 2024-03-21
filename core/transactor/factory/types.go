package factory

// Noncer is an interface for acquiring fresh nonces.
type Noncer interface {
	Acquire() (nonce uint64, isReplacing bool)
}
