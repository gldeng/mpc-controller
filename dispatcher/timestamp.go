package dispatcher

import (
	"strconv"
	"time"
)

func NewTimestamp() string {
	timeString := strconv.Itoa(int(time.Now().UnixMilli()))
	return timeString
}
