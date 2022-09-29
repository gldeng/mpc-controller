package staking

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/tasks/staking/joining"
	"github.com/avalido/mpc-controller/tasks/staking/staking"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/noncer"
)

type Master struct {
	BoundTransactor transactor.Transactor
	DB              storage.DB
	Dispatcher      dispatcher.Dispatcher
	EthClient       chain.EthClient
	Logger          logger.Logger
	NetWorkCtx      chain.NetworkContext
	NonceGiver      noncer.Noncer
	PartiPubKey     storage.PubKey
	SignerMPC       core.Signer
	TxIssuer        txissuer.TxIssuer

	stakeReqAddedH   *joining.StakeRequestAdded
	stakeReqStartedH *staking.Staking
}

func (m *Master) Start(ctx context.Context) error {
	m.subscribe()
	m.stakeReqStartedH.Init(ctx)
	<-ctx.Done()
	return nil
}

func (m *Master) subscribe() {
	stakeAddedHandler := joining.StakeRequestAdded{
		BoundTransactor: m.BoundTransactor,
		DB:              m.DB,
		Logger:          m.Logger,
		PartiPubKey:     m.PartiPubKey,
	}

	taskStartedDealer := staking.Staking{
		BoundTransactor: m.BoundTransactor,
		DB:              m.DB,
		EthClient:       m.EthClient,
		Logger:          m.Logger,
		NetWorkCtx:      m.NetWorkCtx,
		NonceGiver:      m.NonceGiver,
		PartiPubKey:     m.PartiPubKey,
		SignerMPC:       m.SignerMPC,
	}

	m.stakeReqAddedH = &stakeAddedHandler
	m.stakeReqStartedH = &taskStartedDealer

	m.Dispatcher.Subscribe(&events.StakeRequestAdded{}, m.stakeReqAddedH)
	m.Dispatcher.Subscribe(&events.RequestStarted{}, m.stakeReqStartedH)
}
