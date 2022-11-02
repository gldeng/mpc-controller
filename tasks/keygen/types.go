package keygen

const (
	StatusInit Status = iota
	StatusKeygenReqSent
	StatusTxSent
	StatusDone
)

type Status int
