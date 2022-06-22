package dispatcher

import (
	"sync/atomic"
)

var totalEventCount uint64

func SetEventCount(val uint64) {
	atomic.StoreUint64(&totalEventCount, val)
}

func AddEventCount() uint64 {
	return atomic.AddUint64(&totalEventCount, 1)
}

func LoadEventCount() uint64 {
	return atomic.LoadUint64(&totalEventCount)
}
