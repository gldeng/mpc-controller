package joining

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"sync"
	"time"
)

var (
	ErrCannotJoin = errors.New("Cannot join anymore")
)

// Accept event: *contract.MpcManagerStakeRequestAdded

// Emit event:

type StakeRequestAddedEventHandler struct {
	ContractAddr    common.Address
	Logger          logger.Logger
	MyIndexGetter   cache.MyIndexGetter
	MyPubKeyHashHex string
	Publisher       dispatcher.Publisher
	Receipter       chain.Receipter
	Signer          *bind.TransactOpts
	Transactor      bind.ContractTransactor
	evtObjChan      chan *dispatcher.EventObject
	once            sync.Once
}

func (eh *StakeRequestAddedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.evtObjChan = make(chan *dispatcher.EventObject, 1024)
		go eh.joinRequest(ctx)
	})

	select {
	case <-ctx.Done():
		return
	case eh.evtObjChan <- evtObj:
	}
}

func (eh *StakeRequestAddedEventHandler) joinRequest(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-eh.evtObjChan:
			evt, ok := evtObj.Event.(*contract.MpcManagerStakeRequestAdded)
			if !ok {
				break
			}
			genPubKeyHashHex := evt.PublicKey.Hex()
			myIndex := eh.MyIndexGetter.GetMyIndex(eh.MyPubKeyHashHex, genPubKeyHashHex)
			if myIndex == nil {
				break
			}

			txHash, err := eh.doJoinRequest(ctx, myIndex, evt)
			if err != nil {
				if errors.Is(err, ErrCannotJoin) {
					eh.Logger.DebugOnError(err, "Join request unaccepted", []logger.Field{
						{"reqId", evt.RequestId},
						{"txHash", txHash}}...)
					break
				}
				eh.Logger.ErrorOnError(err, "Failed to join request", []logger.Field{
					{"reqId", evt.RequestId},
					{"txHash", txHash}}...)
				break
			}
		}
	}
}

func (eh *StakeRequestAddedEventHandler) doJoinRequest(ctx context.Context, myIndex *big.Int, req *contract.MpcManagerStakeRequestAdded) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
			var err error
			tx, err = transactor.JoinRequest(eh.Signer, req.RequestId, myIndex)
			if err != nil {
				if strings.Contains(err.Error(), "execution reverted: Cannot join anymore") {
					tx = nil
					return false, errors.WithStack(ErrCannotJoin)
				}
				return true, errors.WithStack(err)
			}
			return false, nil
		})

		err = errors.Wrapf(err, "failed to join request")
		if err != nil {
			if errors.Is(err, ErrCannotJoin) {
				return false, err
			}
			return true, err
		}

		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
			rcpt, err := eh.Receipter.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				return true, errors.WithStack(err)
			}

			if rcpt.Status != 1 {
				return true, errors.New("join request tx receipt status != 1")
			}

			return false, nil
		})
		if err != nil {
			return true, errors.WithStack(err)
		}

		return false, nil
	})

	err = errors.Wrapf(err, "failed to join request. reqID:%v, partiIndex:%v", req.RequestId, myIndex)
	if err != nil {
		return
	}
	if tx != nil {
		txHash_ := tx.Hash()
		txHash = &txHash_
		return

	}
	return
}
