package time

import (
	"math/rand"
	"time"
)

func RandomAfter(milliSeconds int64) {
	rand.Seed(time.Now().UnixNano())
	random := rand.Int63n(milliSeconds)
	<-time.After(time.Millisecond * time.Duration(random))
}
