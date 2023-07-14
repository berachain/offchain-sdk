package job

// Result encapsulates the result of a job execution.
type Resultor struct {
	res any
	err error
}

// Result returns the result of the job execution.
func (r *Resultor) Result() any {
	return r.res
}

// Error returns the error of the job execution.
func (r *Resultor) Error() error {
	return r.err
}
