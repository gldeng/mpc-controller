package ethlog

const (
	StatusInit Status = iota
	StatusTxSent
	StatusDone
)

type Status int
