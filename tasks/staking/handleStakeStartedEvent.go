package staking

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
)

type Cache interface {
	cache.MyIndexGetter
	cache.GeneratedPubKeyInfoGetter
	cache.ParticipantKeysGetter
}

// Accept event: *contract.MpcManagerStakeRequestStarted

// Emit event:

type StakeRequestStartedEventHandler struct {
	Logger logger.Logger
	chain.NetworkContext

	MyPubKeyHashHex string

	Cache Cache

	SignDoner core.SignDoner
	Verifyier crypto.VerifyHasher

	Noncer chain.Noncer
	Issuer Issuerer

	genPubKeyInfo *events.GeneratedPubKeyInfo
	myIndex       *big.Int
}

func (eh *StakeRequestStartedEventHandler) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestStarted:
		pubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(evt.PublicKey.Hex())
		if pubKeyInfo == nil {
			eh.Logger.Error("No GeneratedPubKeyInfo found")
			break
		}
		eh.genPubKeyInfo = pubKeyInfo
		index := eh.Cache.GetMyIndex(eh.MyPubKeyHashHex, evt.PublicKey.Hex())
		if pubKeyInfo == nil {
			eh.Logger.Error("Not found my index.")
			break
		}
		eh.myIndex = index

		ok := eh.isParticipant(evtObj.Context, evt)
		if ok {
			nonce, err := eh.getNonce(evtObj.Context)
			if err != err {
				eh.Logger.Error("Failed to get nonce", []logger.Field{{"error", err}}...)
				return
			}

			taskCreator := StakeTaskCreator{
				MpcManagerStakeRequestStarted: evt,
				NetworkContext:                eh.NetworkContext,
				PubKeyHex:                     eh.genPubKeyInfo.GenPubKeyHex,
				Nonce:                         nonce,
			}

			partiKeys, err := eh.getNormalizedPartiKeys(evtObj.Context, evt.PublicKey, evt.ParticipantIndices)
			if err != nil {
				eh.Logger.Error("Failed to get normalized participant keys", []logger.Field{{"error", err}}...)
				return
			}

			signReqCreator := SignRequestCreator{
				TaskID:                    evt.Raw.TxHash.Hex(),
				NormalizedParticipantKeys: partiKeys,
				PubKeyHex:                 eh.genPubKeyInfo.GenPubKeyHex,
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

func (eh *StakeRequestStartedEventHandler) isParticipant(ctx context.Context, req *contract.MpcManagerStakeRequestStarted) bool {
	var participating bool
	for _, index := range req.ParticipantIndices {
		if index.Cmp(eh.myIndex) == 0 {
			participating = true
			break
		}
	}

	if !participating {
		eh.Logger.Info("Not participated to stake request", []logger.Field{{"stakeReqId", req.RequestId}}...)
		return false
	}

	return true
}

func (eh *StakeRequestStartedEventHandler) getNonce(ctx context.Context) (uint64, error) {
	pubkey, err := crypto.UnmarshalPubKeyHex(eh.genPubKeyInfo.GenPubKeyHex)
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

func (eh *StakeRequestStartedEventHandler) getNormalizedPartiKeys(ctx context.Context, genPubKeyHash common.Hash, partiIndices []*big.Int) ([]string, error) {
	partiKeys := eh.Cache.GetParticipantKeys(genPubKeyHash, partiIndices)
	if partiKeys == nil {
		return nil, errors.New("Found no participant keys.")
	}

	normalized, err := crypto.NormalizePubKeys(partiKeys)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to normalized participant public keys: %v", partiKeys)
	}

	return normalized, nil
}
