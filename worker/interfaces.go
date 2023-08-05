package worker

// Payload encapsulates a job and its input into a neat package to
// be executed by another thread.
type Payload interface {
	Execute() Resultor
}

// Resultor encapsulates the result of a job execution.
type Resultor interface {
	Result() any
	Error() error
}
