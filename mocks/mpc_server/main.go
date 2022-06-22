package main

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mocks/mpc_server/service"
)

func main() {
	logger.DevMode = true
	service.ListenAndServe("9000")
}
