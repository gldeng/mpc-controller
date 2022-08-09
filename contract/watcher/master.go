package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/common"
)

type Master struct {
	BoundCaller     caller.Caller
	BoundTransactor transactor.Transactor
	ContractAddr    common.Address
	DB              storage.DB
	Dispatcher      dispatcher.Dispatcher
	EthWsURL        string
	KeyGeneratorMPC core.KeygenDoner
	Logger          logger.Logger
	PartiPubKey     storage.PubKey

	watcher *MpcManagerWatchers
}

func (m *Master) Start(ctx context.Context) error {
	m.subscribe()
	m.watcher.Init(ctx)
	<-ctx.Done()
	return nil
}

func (m *Master) subscribe() {
	watcher := MpcManagerWatchers{
		BoundCaller:     m.BoundCaller,
		BoundTransactor: m.BoundTransactor,
		ContractAddr:    m.ContractAddr,
		DB:              m.DB,
		EthWsURL:        m.EthWsURL,
		KeyGeneratorMPC: m.KeyGeneratorMPC,
		Logger:          m.Logger,
		PartiPubKey:     m.PartiPubKey,
		Publisher:       m.Dispatcher,
	}

	m.watcher = &watcher

	m.Dispatcher.Subscribe(&events.ParticipantAdded{}, m.watcher)
	m.Dispatcher.Subscribe(&events.KeyGenerated{}, m.watcher)
}
