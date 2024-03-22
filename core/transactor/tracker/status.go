package tracker

// Status represents the current status of a tx owned by the transactor. These are used only after
// the tx status has been confirmed by erroring, the chain, or the configured timeout.
type Status uint8

const (
	StatusPending Status = iota
	StatusError
	StatusSuccess
	StatusReverted
	StatusStale
)
