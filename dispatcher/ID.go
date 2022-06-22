package dispatcher

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"time"
)

func NewID() string {
	ts := time.Now().Format(time.RFC3339Nano)
	us := uuid.NewString()
	content := ts + us

	hasher := md5.New()
	hasher.Write([]byte(content))
	return hex.EncodeToString(hasher.Sum(nil))[:8]
}
