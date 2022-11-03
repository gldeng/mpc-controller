package join

const (
	StatusInit Status = iota
	StatusTxSent
	StatusDone
)

type Status int
