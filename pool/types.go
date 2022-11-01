package pool

// TODO: Deprecate
type WorkerPool interface {
	Submit(task func())
	StopAndWait()
}
