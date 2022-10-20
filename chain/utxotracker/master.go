package utxotracker

import (
	"context"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	kbcevents "github.com/kubecost/events"
)

type Master struct {
	Ctx                     context.Context
	ClientPChain            platformvm.Client
	Dispatcher              dispatcher.Dispatcher
	UTXOToRecoverDispatcher kbcevents.Dispatcher[*events.UTXOToRecover]
	Logger                  logger.Logger
	utxoTracker             *UTXOTracker
}

func (m *Master) Start() error {
	m.subscribe()
	<-m.Ctx.Done()
	return nil
}

func (m *Master) subscribe() {
	utxoTracker := UTXOTracker{
		ClientPChain: m.ClientPChain,
		Logger:       m.Logger,
		Dispatcher:   m.UTXOToRecoverDispatcher,
	}

	m.utxoTracker = &utxoTracker

	m.Dispatcher.Subscribe(&events.KeyGenerated{}, m.utxoTracker)
}

func (m *Master) Close() error {
	// todo
	return nil
}
