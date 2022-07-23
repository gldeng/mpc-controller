package service

import (
	"github.com/avalido/mpc-controller/logger"
	"testing"
)

func TestListenAndServe(t *testing.T) {
	logger.DevMode = true
	ListenAndServe("9000")
}
