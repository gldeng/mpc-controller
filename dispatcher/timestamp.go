package dispatcher

import (
	"strconv"
	"time"
)

func NowTimestamp() string {
	timeString := strconv.Itoa(int(time.Now().UnixMilli()))
	return timeString
}
