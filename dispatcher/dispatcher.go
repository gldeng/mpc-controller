package dispatcher

type EventDispatcher interface {
	Dispatch(event any)
}
