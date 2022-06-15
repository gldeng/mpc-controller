package staking

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/pkg/errors"
	"math/big"
)

type Storer interface {
	storage.StorerLoadGeneratedPubKeyInfo
	storage.StorerLoadParticipantInfo
	storage.StorerGetPariticipantKeys
}

// Accept event: *contract.MpcManagerStakeRequestStarted

// Emit event:

type StakeRequestStartedEventHandler struct {
	Logger logger.Logger
	chain.NetworkContext

	MyPubKeyHashHex string

	Storer Storer

	SignDoner core.SignDoner
	Verifyier crypto.VerifyHasher

	Noncer chain.Noncer
	Issuer Issuerer

	genPubKeyInfo *storage.GeneratedPubKeyInfo
}

func (eh *StakeRequestStartedEventHandler) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestStarted:
		ok, err := eh.isParticipant(evtObj.Context, evt)
		eh.Logger.ErrorOnError(err, "Failed to check participant", []logger.Field{{"error", err}}...)
		if ok {
			nonce, err := eh.getNonce(evtObj.Context)
			if err != err {
				eh.Logger.Error("Failed to get nonce", []logger.Field{{"error", err}}...)
				return
			}

			taskCreator := StakeTaskCreator{
				MpcManagerStakeRequestStarted: evt,
				NetworkContext:                eh.NetworkContext,
				PubKeyHex:                     eh.genPubKeyInfo.PubKeyHex,
				Nonce:                         nonce,
			}

			partiKeys, err := eh.getNormalizedPartiKeys(evtObj.Context, evt.PublicKey.Hex(), evt.ParticipantIndices)
			if err != nil {
				eh.Logger.Error("Failed to get normalized participant keys", []logger.Field{{"error", err}}...)
				return
			}

			signReqCreator := SignRequestCreator{
				TaskID:                    evt.Raw.TxHash.Hex(),
				NormalizedParticipantKeys: partiKeys,
				PubKeyHex:                 eh.genPubKeyInfo.PubKeyHex,
			}

			taskSignRequester := StakeTaskSignRequester{
				StakeTaskCreatorer:   &taskCreator,
				SignRequestCreatorer: &signReqCreator,
				SignDoner:            eh.SignDoner,
				VerifyHasher:         eh.Verifyier,
			}

			task, err := taskSignRequester.Sign(evtObj.Context)
			if err != nil {
				eh.Logger.Error("Failed to sign stake task", []logger.Field{{"error", err}, {"reqID", evt.RequestId}}...)
				return
			}

			_, err = eh.Issuer.IssueTask(evtObj.Context, task)
			if err != nil {
				eh.Logger.Error("Failed to issue stake task", []logger.Field{{"error", err}, {"reqID", evt.RequestId}}...)
			}
			eh.Logger.Info("Cool! Success to add delegator!", []logger.Field{{"stakeTaske", task}}...)
		}
	}
}

func (eh *StakeRequestStartedEventHandler) isParticipant(ctx context.Context, req *contract.MpcManagerStakeRequestStarted) (bool, error) {
	myIndex, err := eh.getMyIndex(ctx, req.PublicKey.Hex())
	if err != nil {
		return false, errors.WithStack(err)
	}

	var participating bool
	for _, index := range req.ParticipantIndices {
		if index.Cmp(myIndex) == 0 {
			participating = true
			break
		}
	}

	if !participating {
		eh.Logger.Info("Not participated to stake request", []logger.Field{{"stakeReqId", req.RequestId}}...)
		return false, nil
	}

	return true, nil
}

func (eh *StakeRequestStartedEventHandler) getMyIndex(ctx context.Context, genPubKeyHashHex string) (*big.Int, error) {
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

func (eh *StakeRequestStartedEventHandler) getNonce(ctx context.Context) (uint64, error) {
	pubkey, err := crypto.UnmarshalPubKeyHex(eh.genPubKeyInfo.PubKeyHex)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	address := crypto.PubkeyToAddresse(pubkey)

	nonce, err := eh.Noncer.NonceAt(ctx, *address, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return nonce, nil
}

func (eh *StakeRequestStartedEventHandler) getNormalizedPartiKeys(ctx context.Context, pubKeyHex string, partiIndices []*big.Int) ([]string, error) {
	partiKeys, err := eh.getNormalizedPartiKeys(ctx, pubKeyHex, partiIndices)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	normalized, err := crypto.NormalizePubKeys(partiKeys)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to normalized participant public keys: %v", partiKeys)
	}

	return normalized, nil
}
