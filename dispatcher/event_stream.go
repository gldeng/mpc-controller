package dispatcher

import (
	"sync/atomic"
)

var totalEventStreamCount uint64

func SetEventStreamCount(val uint64) {
	atomic.StoreUint64(&totalEventStreamCount, val)
}

func AddEventStreamCount() uint64 {
	return atomic.AddUint64(&totalEventStreamCount, 1)
}

func LoadEventStreamCount() uint64 {
	return atomic.LoadUint64(&totalEventStreamCount)
}
