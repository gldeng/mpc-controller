package keygen

const (
	StatusInit Status = iota
	StatusKeygenReqSent
	StatusTxSent
	StatusDone

	StatusDropped
)

type Status int
