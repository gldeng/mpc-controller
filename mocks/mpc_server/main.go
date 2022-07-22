package main

import (
	"flag"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mocks/mpc_server/service"
)

func main() {
	var participants = flag.Int("p", 7, "number of participants in the group")
	var threshold = flag.Int("t", 4, "number of the group threshold")
	flag.Parse()

	logger.DevMode = true
	service.ListenAndServe("9000", *participants, *threshold)
}
