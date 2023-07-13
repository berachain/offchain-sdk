package jobs

type BasicJob interface {
	Execute() error
}

type ConditionalJob interface {
	BasicJob
	Condition() bool
}
