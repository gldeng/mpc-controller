package events

type MpcServerResultFetched struct {
	ReqID     string
	Result    string
	ReqType   string
	ReqStatus string
}
