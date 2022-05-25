package main

import "github.com/avalido/mpc-controller/mocks/mpc_server/service"

func main() {
	service.ListenAndServe("9000")
}
