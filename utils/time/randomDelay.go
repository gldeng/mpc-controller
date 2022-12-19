package time

import (
	"math/rand"
	"time"
)

func RandomDelay(maxMilliSeconds int64) {
	rand.Seed(time.Now().UnixNano())
	random := rand.Int63n(maxMilliSeconds)
	<-time.After(time.Millisecond * time.Duration(random))
}
