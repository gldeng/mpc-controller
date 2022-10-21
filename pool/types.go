package pool

import (
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/noncer"
)

// TODO: Deprecate
type WorkerPool interface {
	Submit(task func())
	StopAndWait()
}

type Task interface {
	Next(resources *Resources) ([]Task, error)
}

type Resources struct { // TODO: Convert it to TaskApi interface instead of directly giving the underlying resources
	Logger logger.Logger

	NonceGiver noncer.Noncer
	Network    chain.NetworkContext

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer
}

type ResourcesFactory = func() *Resources
