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
)

type JoiningMaster struct {
	Logger       logger.Logger
	ContractAddr common.Address

	MyPubKeyHashHex string
	MyIndexGetter   cache.MyIndexGetter

	Dispatcher dispatcher.DispatcherClaasic

	Signer *bind.TransactOpts

	Receipter chain.Receipter

	Transactor bind.ContractTransactor

	joiningWatcher *StakeRequestAddedEventWatcher
	joiningDealer  *StakeRequestAddedEventHandler
}

func (j *JoiningMaster) Start(ctx context.Context) error {
	j.subscribe()
	<-ctx.Done()
	return nil
}

func (j *JoiningMaster) subscribe() {
	stakeAddedWatcher := StakeRequestAddedEventWatcher{
		Logger:       j.Logger,
		ContractAddr: j.ContractAddr,
		Publisher:    j.Dispatcher,
	}

	stakeAddedHandler := StakeRequestAddedEventHandler{
		Logger:          j.Logger,
		MyPubKeyHashHex: j.MyPubKeyHashHex,
		Signer:          j.Signer,
		MyIndexGetter:   j.MyIndexGetter,
		Receipter:       j.Receipter,
		ContractAddr:    j.ContractAddr,
		Transactor:      j.Transactor,
		Publisher:       j.Dispatcher,
	}

	j.joiningWatcher = &stakeAddedWatcher
	j.joiningDealer = &stakeAddedHandler

	j.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, j.joiningWatcher)
	j.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, j.joiningWatcher) // Emit event:  *contract.MpcManagerStakeRequestAdded
	j.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestAdded{}, j.joiningDealer)   // Emit event: *events.JoinedRequestEvent
}
