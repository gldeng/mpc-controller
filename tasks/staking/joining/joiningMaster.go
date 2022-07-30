package joining

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

type JoiningMaster struct {
	ContractAddr       common.Address
	Dispatcher         dispatcher.DispatcherClaasic
	Logger             logger.Logger
	MyIndexGetter      cache.MyIndexGetter
	MyPubKeyHashHex    string
	Receipter          chain.Receipter
	Signer             *bind.TransactOpts
	StakeReqCacheCap   uint32
	StakeReqPublishDur time.Duration
	Transactor         bind.ContractTransactor

	joiningWatcher *StakeRequestAddedEventWatcher
	joiningDealer  *StakeRequestAddedEventHandler
}

func (m *JoiningMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *JoiningMaster) subscribe() {
	stakeAddedWatcher := StakeRequestAddedEventWatcher{
		Logger:       m.Logger,
		ContractAddr: m.ContractAddr,
		Publisher:    m.Dispatcher,
	}

	stakeAddedHandler := StakeRequestAddedEventHandler{
		Logger:          m.Logger,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		Signer:          m.Signer,
		MyIndexGetter:   m.MyIndexGetter,
		Receipter:       m.Receipter,
		ContractAddr:    m.ContractAddr,
		Transactor:      m.Transactor,
		Publisher:       m.Dispatcher,
	}

	m.joiningWatcher = &stakeAddedWatcher
	m.joiningDealer = &stakeAddedHandler

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.joiningWatcher)
	m.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, m.joiningWatcher) // Emit event:  *contract.MpcManagerStakeRequestAdded
	m.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestAdded{}, m.joiningDealer)   // Emit event: *events.JoinedRequestEvent
}
