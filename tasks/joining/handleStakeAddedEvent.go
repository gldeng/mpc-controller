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

// Accept event: *contract.MpcManagerStakeRequestAdded

// Emit event: *events.JoinedRequestEvent

type StakeRequestAddedEventHandler struct {
	Logger logger.Logger

	MyPubKeyHashHex string
	MyIndexGetter   cache.MyIndexGetter

	Signer *bind.TransactOpts

	Receipter chain.Receipter

	ContractAddr common.Address
	Transactor   bind.ContractTransactor

	Publisher dispatcher.Publisher
}

func (eh *StakeRequestAddedEventHandler) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestAdded:
		genPubKeyHashHex := evt.PublicKey.Hex()
		myIndex := eh.MyIndexGetter.GetMyIndex(eh.MyPubKeyHashHex, genPubKeyHashHex)
		if myIndex == nil {
			break
		}
		dur := rand.Intn(5000)
		time.Sleep(time.Millisecond * time.Duration(dur)) // sleep because concurrent joinRequest can cause failure.
		txHash, err := eh.joinRequest(evtObj.Context, myIndex, evt)
		if err != nil {
			eh.Logger.Error("Failed to join request", []logger.Field{
				{"error", err},
				{"reqId", evt.RequestId},
				{"txHash", txHash}}...)
			break
		}

		if txHash != nil {
			newEvt := events.JoinedRequestEvent{
				TxHashHex: txHash.Hex(),
				RequestId: evt.RequestId,
				Index:     myIndex,
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

	err = backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		var err error
		tx, err = transactor.JoinRequest(eh.Signer, req.RequestId, myIndex)
		if err != nil {
			if strings.Contains(err.Error(), "execution reverted: Cannot join anymore") {
				tx = nil
				eh.Logger.Info("Cannot join anymore", []logger.Field{{"reqId", req.RequestId}, {"myIndex", myIndex}}...)
				return nil
			}
			return errors.Wrapf(err, "failed to join request. ReqId: %v, Index: %v", req.RequestId, myIndex)
		}

		time.Sleep(time.Second * 3)

		rcpt, err := eh.Receipter.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			return errors.New("Transaction failed")
		}

		return nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if tx != nil {
		txHash_ := tx.Hash()
		return &txHash_, nil

	}

	return nil, err
}
