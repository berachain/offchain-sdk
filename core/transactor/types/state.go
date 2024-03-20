package types

// PreconfirmedStates are used for messages before the tx status is confirmed by the chain.
type PreconfirmedState uint8

const (
	// The message is not being tracked by the transactor.
	StateUnknown PreconfirmedState = iota
	// The message is sitting in the queue, waiting to be acquired into a tx.
	StateQueued
	// The message is being built into a tx.
	StateBuilding
	// The tx containing the message is sending (or retrying) -- noncer marked as "acquired".
	StateSending
	// The tx containing the message has been sent -- noncer marked as "inFlight".
	StateInFlight
)
