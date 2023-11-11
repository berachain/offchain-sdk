package job

type WorkerPool interface {
	Submit(func())
	SubmitAndWait(func())
}
