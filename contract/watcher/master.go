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
	kbcevents "github.com/kubecost/events"
)

type Master struct {
	Ctx                     context.Context
	BoundCaller             caller.Caller
	BoundTransactor         transactor.Transactor
	ContractAddr            common.Address
	DB                      storage.DB
	Dispatcher              dispatcher.Dispatcher // TODO: use kubecost/events instead
	StakeReqAddedDispatcher kbcevents.Dispatcher[*events.StakeRequestAdded]
	ReqStartedDispatcher    kbcevents.Dispatcher[*events.RequestStarted]
	EthWsURL                string
	MpcClient               core.MpcClient
	Logger                  logger.Logger
	PartiPubKey             storage.PubKey

	watcher *MpcManagerWatchers
}

func (m *Master) Start() error {
	m.subscribe()
	m.watcher.Init(m.Ctx)
	return nil
}

func (m *Master) Close() error {
	// todo:
	return nil
}

func (m *Master) subscribe() {
	watcher := MpcManagerWatchers{
		BoundCaller:             m.BoundCaller,
		BoundTransactor:         m.BoundTransactor,
		ContractAddr:            m.ContractAddr,
		DB:                      m.DB,
		EthWsURL:                m.EthWsURL,
		MpcClient:               m.MpcClient,
		Logger:                  m.Logger,
		PartiPubKey:             m.PartiPubKey,
		Publisher:               m.Dispatcher,
		StakeReqAddedDispatcher: m.StakeReqAddedDispatcher,
		ReqStartedDispatcher:    m.ReqStartedDispatcher,
	}

	m.watcher = &watcher

	m.Dispatcher.Subscribe(&events.ParticipantAdded{}, m.watcher)
	m.Dispatcher.Subscribe(&events.KeyGenerated{}, m.watcher)
}
