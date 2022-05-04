package mpc_client

import (
	"sync"
)

var rmLock = &sync.RWMutex{}

var reqNumMapCache map[string]map[RequestType]int

func init() {
	reqNumMapCache = make(map[string]map[RequestType]int)
}

func IncreaseRequestNumber(reqId string, reqType RequestType) {
	rmLock.Lock()
	defer rmLock.Unlock()

	_, ok := reqNumMapCache[reqId]
	if !ok {
		reqNumMapCache[reqId] = map[RequestType]int{reqType: 1}
		return
	}
	reqNumMapCache[reqId][reqType]++
}

func QueryRequestNumber(reqId string, reqType RequestType) int {
	rmLock.RLock()
	defer rmLock.RUnlock()

	return reqNumMapCache[reqId][reqType]
}
