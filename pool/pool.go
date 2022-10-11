package pool

type Task func()

type WorkerPool interface {
	Submit(task Task)
	StopAndWait()
}
