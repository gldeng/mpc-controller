package joining

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

var (
	ErrCannotJoin = errors.New("Cannot join anymore")
)

// Accept event: *contract.MpcManagerStakeRequestAdded

// Emit event: *events.JoinedRequestEvent

type StakeRequestAddedEventHandler struct {
	ContractAddr    common.Address
	Logger          logger.Logger
	MyIndexGetter   cache.MyIndexGetter
	MyPubKeyHashHex string
	Publisher       dispatcher.Publisher
	Receipter       chain.Receipter
	Signer          *bind.TransactOpts
	Transactor      bind.ContractTransactor
}

func (eh *StakeRequestAddedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestAdded:
		genPubKeyHashHex := evt.PublicKey.Hex()
		myIndex := eh.MyIndexGetter.GetMyIndex(eh.MyPubKeyHashHex, genPubKeyHashHex)
		if myIndex == nil {
			break
		}

		n := rand.Intn(10_000_000_000)
		time.Sleep(time.Nanosecond * time.Duration(n)) // reduce concurrent conflict

		txHash, err := eh.joinRequest(evtObj.Context, myIndex, evt)
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

		if txHash != nil {
			newEvt := events.JoinedRequestEvent{
				TxHashHex:  txHash.Hex(),
				RequestId:  evt.RequestId,
				PartiIndex: myIndex,
			}

			eh.Publisher.Publish(evtObj.Context, dispatcher.NewEventObjectFromParent(evtObj, "StakeRequestAddedEventHandler", &newEvt, evtObj.Context))
		}
	}
}

func (eh *StakeRequestAddedEventHandler) joinRequest(ctx context.Context, myIndex *big.Int, req *contract.MpcManagerStakeRequestAdded) (txHash *common.Hash, err error) {
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

		time.Sleep(time.Second * 5)

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
