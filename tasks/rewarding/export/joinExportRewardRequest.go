package export

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

// Accept event: *events.ExportRewardRequestAddedEvent

// Emit event: *events.JoinedExportRewardRequestEvent

type ExportRewardRequestJoiner struct {
	Logger logger.Logger

	MyPubKeyHashHex string

	Signer *bind.TransactOpts
	Cache  cache.ICache

	Receipter chain.Receipter

	ContractAddr common.Address
	Transactor   bind.ContractTransactor

	Publisher dispatcher.Publisher
}

func (eh *ExportRewardRequestJoiner) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ExportRewardRequestAddedEvent:
		genPubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(evt.PublicKeyHash.Hex())
		if genPubKeyInfo == nil {
			break
		}

		myIndex := eh.Cache.GetMyIndex(eh.MyPubKeyHashHex, evt.PublicKeyHash.Hex())
		if myIndex == nil {
			break
		}

		groupIDBytes := bytes.HexTo32Bytes(genPubKeyInfo.GroupIdHex)
		pubKeyBytes := bytes.HexToBytes(genPubKeyInfo.GenPubKeyHex)

		dur := rand.Intn(5000)
		time.Sleep(time.Millisecond * time.Duration(dur)) // sleep because concurrent joinExportRewardRequest can cause failure.
		txHash, err := eh.joinExportRewardRequest(ctx, groupIDBytes, myIndex, pubKeyBytes, evt.AddDelegatorTxID)
		if err != nil {
			eh.Logger.Error("Failed to join export reward request", []logger.Field{
				{"error", err},
				{"addDelegatorTxID", bytes.Bytes32ToHex(evt.AddDelegatorTxID)},
				{"txHash", txHash}}...)
			break
		}

		if txHash != nil {
			newEvt := events.JoinedExportRewardRequestEvent{
				GroupIDHex:       genPubKeyInfo.GroupIdHex,
				MyIndex:          myIndex,
				PubKeyHex:        genPubKeyInfo.GenPubKeyHex,
				AddDelegatorTxID: evt.AddDelegatorTxID,
				TxHash:           *txHash,
			}

			eh.Publisher.Publish(evtObj.Context, dispatcher.NewEventObjectFromParent(evtObj, "ExportRewardRequestJoiner", &newEvt, evtObj.Context))
		}
	}
}

func (eh *ExportRewardRequestJoiner) joinExportRewardRequest(ctx context.Context, groupId [32]byte, myIndex *big.Int, publicKey []byte, txID [32]byte) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction

	err = backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		var err error
		tx, err = transactor.JoinExportReward(eh.Signer, groupId, myIndex, publicKey, txID)
		if err != nil {
			if strings.Contains(err.Error(), "execution reverted: Cannot join anymore") {
				tx = nil
				eh.Logger.Info("Cannot join anymore", []logger.Field{{"addDelegatorTxID", bytes.Bytes32ToHex(txID)}, {"myIndex", myIndex}}...)
				return nil
			}
			return errors.Wrapf(err, "failed to join request. addDelegatorTxID: %v, Index: %v", bytes.Bytes32ToHex(txID), myIndex)
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
