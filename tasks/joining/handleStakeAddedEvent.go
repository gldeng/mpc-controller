package joining

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

type Storer interface {
	storage.StorerLoadGeneratedPubKeyInfo
	storage.StorerLoadParticipantInfo
}

// Accept event: *contract.MpcManagerStakeRequestAdded

// Emit event: *events.JoinedRequestEvent

type StakeRequestAddedEventHandler struct {
	Logger logger.Logger

	MyPubKeyHashHex string

	Signer *bind.TransactOpts

	Storer Storer

	Receipter chain.Receipter

	ContractAddr common.Address
	Transactor   bind.ContractTransactor

	Publisher dispatcher.Publisher

	genPubKeyInfo *storage.GeneratedPubKeyInfo
}

func (eh *StakeRequestAddedEventHandler) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestAdded:
		myIndex, err := eh.getMyIndex(evtObj.Context, evt.PublicKey.Hex())
		if err != nil {
			eh.Logger.Error("Failed to get my index", []logger.Field{
				{"error", err},
				{"pubKeyHashHex", evt.PublicKey.Hex()}}...)
			break
		}

		txHash, err := eh.joinRequest(evtObj.Context, myIndex, evt)
		if err != nil || eh.checkReceipt(*txHash) != nil {
			eh.Logger.Error("Failed to join request", []logger.Field{
				{"error", err},
				{"reqId", evt.RequestId},
				{"txHash", txHash}}...)
			break
		}

		newEvt := events.JoinedRequestEvent{
			TxHashHex: txHash.Hex(),
			RequestId: evt.RequestId,
			Index:     myIndex,
		}

		if err == nil {
			eh.Publisher.Publish(evtObj.Context, dispatcher.NewEventObjectFromParent(evtObj, "StakeRequestAddedEventHandler", newEvt, evtObj.Context))
		}
	}
}

func (eh *StakeRequestAddedEventHandler) joinRequest(ctx context.Context, myIndex *big.Int, req *contract.MpcManagerStakeRequestAdded) (*common.Hash, error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction

	err = backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		var err error
		tx, err = transactor.JoinRequest(eh.Signer, req.RequestId, myIndex)
		if err != nil {
			return errors.Wrapf(err, "failed to join request. ReqId: %v, Index: %v", req.RequestId, myIndex)
		}

		return nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	txHash := tx.Hash()
	return &txHash, nil
}

func (eh *StakeRequestAddedEventHandler) checkReceipt(txHash common.Hash) error {
	rCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := backoff.RetryFnExponentialForever(eh.Logger, rCtx, func() error {

		rcpt, err := eh.Receipter.TransactionReceipt(rCtx, txHash)
		if err != nil {
			return errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			return errors.New("Transaction failed")
		}

		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "transaction %q failed", txHash)
	}

	return nil
}

func (eh *StakeRequestAddedEventHandler) getMyIndex(ctx context.Context, genPubKeyHashHex string) (*big.Int, error) {
	genPubKeyInfo, err := eh.Storer.LoadGeneratedPubKeyInfo(ctx, genPubKeyHashHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	eh.genPubKeyInfo = genPubKeyInfo
	partInfo, err := eh.Storer.LoadParticipantInfo(ctx, eh.MyPubKeyHashHex, genPubKeyInfo.GroupIdHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return big.NewInt(int64(partInfo.Index)), nil
}
