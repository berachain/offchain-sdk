package types

// PreconfirmStates are used for messages before the tx status is confirmed by the chain.
type PreconfirmState uint8

const (
	// The message is not being tracked by the transactor.
	StateUnknown PreconfirmState = iota
	// The message is sitting in the queue, waiting to be acquired into a tx.
	StateQueued
	// The message is being built into a tx.
	StateBuilding
	// The tx containing msg is sending (or retrying), noncer has "acquired".
	StateSending
	// The tx containing the message has been sent, noncer marks as "inFlight".
	StateInFlight
)
