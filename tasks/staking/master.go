package staking

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/staking/joining"
	"github.com/avalido/mpc-controller/tasks/staking/staking"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/common"
)

type StakingMaster struct {
	Balancer          chain.Balancer
	CChainIssueClient chain.CChainIssuer
	Cache             staking.Cache
	ChainNoncer       chain.Noncer
	ContractAddr      common.Address
	Dispatcher        dispatcher.Dispatcher
	Logger            logger.Logger
	MyPubKeyHashHex   string
	Noncer            noncer.Noncer
	PChainIssueClient chain.PChainIssuer
	SignDoner         core.SignDoner
	Transactor        transactor.Transactor
	chain.NetworkContext

	joiningDealer *joining.StakeRequestAdded
	stakingDealer *staking.StakeRequestStarted
}

func (m *StakingMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *StakingMaster) subscribe() {
	stakeAddedHandler := joining.StakeRequestAdded{
		Logger:     m.Logger,
		Transactor: m.Transactor,
	}

	taskStartedDealer := staking.StakeRequestStarted{
		Balancer:          m.Balancer,
		CChainIssueClient: m.CChainIssueClient,
		Cache:             m.Cache,
		ChainNoncer:       m.ChainNoncer,
		Logger:            m.Logger,
		NetworkContext:    m.NetworkContext,
		Noncer:            m.Noncer,
		PChainIssueClient: m.PChainIssueClient,
		Publisher:         m.Dispatcher,
		SignDoner:         m.SignDoner,
	}

	m.joiningDealer = &stakeAddedHandler
	m.stakingDealer = &taskStartedDealer

	m.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestAdded{}, m.joiningDealer)
	m.Dispatcher.Subscribe(&events.RequestStarted{}, m.stakingDealer)
}
