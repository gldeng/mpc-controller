package pool

type WorkerPool interface {
	Submit(task func( ())
	StopAndWait()
}
