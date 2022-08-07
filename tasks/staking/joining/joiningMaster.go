package joining

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

type JoiningMaster struct {
	ContractAddr       common.Address
	Dispatcher         dispatcher.Dispatcher
	Logger             logger.Logger
	MyIndexGetter      cache.MyIndexGetter
	MyPubKeyHashHex    string
	Receipter          chain.Receipter
	Signer             *bind.TransactOpts
	StakeReqCacheCap   uint32
	StakeReqPublishDur time.Duration
	Transactor         transactor.Transactor

	joiningDealer *StakeRequestAdded
}

func (m *JoiningMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *JoiningMaster) subscribe() {
	stakeAddedHandler := StakeRequestAdded{
		Logger:     m.Logger,
		Transactor: m.Transactor,
	}

	m.joiningDealer = &stakeAddedHandler

	m.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestAdded{}, m.joiningDealer)
}
