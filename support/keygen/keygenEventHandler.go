package keygen

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/ids"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
	"time"
)

// Accept event: *events.GroupInfoStoredEvent
// Accept event: *events.ParticipantInfoStoredEvent
// Accept event: *contract.MpcManagerKeygenRequestAdded

// Emit event: *events.GeneratedPubKeyInfoStoredEvent
// Emit event: *events.ReportedGenPubKeyEvent

type KeygenRequestAddedEventHandler struct {
	ContractAddr    common.Address
	KeygenDoner     core.KeygenDoner
	Logger          logger.Logger
	MyPubKeyHashHex string
	Publisher       dispatcher.Publisher
	Receipter       chain.Receipter
	Signer          *bind.TransactOpts
	Storer          storage.MarshalSetter
	Transactor      bind.ContractTransactor

	groupInfoMap       map[string]events.GroupInfo
	participantInfoMap map[string]events.ParticipantInfo
}

// Pre-condition: *contract.MpcManagerKeygenRequestAdded must happen after *event.GroupInfoStoredEvent

func (eh *KeygenRequestAddedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.GroupInfoStoredEvent:
		if len(eh.groupInfoMap) == 0 {
			eh.groupInfoMap = make(map[string]events.GroupInfo)
		}
		eh.groupInfoMap[evt.Key] = evt.Val
	case *events.ParticipantInfoStoredEvent:
		if len(eh.participantInfoMap) == 0 {
			eh.participantInfoMap = make(map[string]events.ParticipantInfo)
		}
		eh.participantInfoMap[evt.Key] = evt.Val
	case *contract.MpcManagerKeygenRequestAdded:
		err := eh.do(evtObj.Context, evt, evtObj)
		eh.Logger.ErrorOnError(err, "Failed to deal with MpcManagerKeygenRequestAdded event.", []logger.Field{{"error", err}}...)
	}
}

func (eh *KeygenRequestAddedEventHandler) do(ctx context.Context, req *contract.MpcManagerKeygenRequestAdded, evtObj *dispatcher.EventObject) error {
	reqId := req.Raw.TxHash.Hex()

	groupIdHex := bytes.Bytes32ToHex(req.GroupId)
	group := eh.groupInfoMap[events.PrefixGroupInfo+"-"+groupIdHex]

	keyGenReq := &core.KeygenRequest{
		RequestId:       reqId,
		ParticipantKeys: group.PartPubKeyHexs,
		Threshold:       group.Threshold,
	}

	genPubKeyHex, err := eh.keygen(ctx, keyGenReq)
	if err != nil {
		return errors.WithStack(err)
	}

	cchainAddr, err := addrs.PubKeyHexToAddress(genPubKeyHex)
	if err != nil {
		return errors.WithStack(err)
	}

	pchainAddr, err := ids.ShortIDFromPubKeyHex(genPubKeyHex)
	if err != nil {
		return errors.WithStack(err)
	}

	dnmGenPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(genPubKeyHex) // for Ethereum compatibility
	if err != nil {
		return errors.WithStack(err)
	}
	myIndex := eh.participantInfoMap[events.PrefixParticipantInfo+"-"+eh.MyPubKeyHashHex+"-"+groupIdHex].Index
	dur := rand.Intn(5000)
	time.Sleep(time.Millisecond * time.Duration(dur)) // sleep because concurrent reporting can cause failure.
	err = eh.reportGeneratedKey(evtObj.Context, req.GroupId, big.NewInt(int64(myIndex)), dnmGenPubKeyBytes)
	if err != nil {
		return errors.WithStack(err)
	}

	dnmGenPubKeyHash := hash256.FromHex(bytes.BytesToHex(dnmGenPubKeyBytes))

	reportedEvt := events.ReportedGenPubKeyEvent{
		GroupIdHex:       groupIdHex,
		PartiIndex:       big.NewInt(int64(myIndex)),
		GenPubKeyHex:     bytes.BytesToHex(dnmGenPubKeyBytes),
		GenPubKeyHashHex: dnmGenPubKeyHash.Hex(),
		CChainAddress:    *cchainAddr,
		PChainAddress:    *pchainAddr,
	}

	eh.publishReportedEvent(ctx, &reportedEvt, evtObj)

	genPubKeyInfo := GeneratedPubKeyInfo{
		GenPubKeyHashHex:       dnmGenPubKeyHash.Hex(),
		CompressedGenPubKeyHex: genPubKeyHex,
		GroupIdHex:             groupIdHex,
	}

	key, err := eh.storeGenKenInfo(ctx, &genPubKeyInfo)
	if err != nil {
		return errors.WithStack(err)
	}

	eh.publishStoredEvent(ctx, key, &genPubKeyInfo, evtObj)
	return nil
}

func (eh *KeygenRequestAddedEventHandler) keygen(ctx context.Context, req *core.KeygenRequest) (string, error) {
	res, err := eh.KeygenDoner.KeygenDone(ctx, req)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return res.Result, nil
}

func (eh *KeygenRequestAddedEventHandler) storeGenKenInfo(ctx context.Context, genkenInfo *GeneratedPubKeyInfo) (string, error) {
	key := prefixGeneratedPubKeyInfo + "-" + genkenInfo.GenPubKeyHashHex

	err := eh.Storer.MarshalSet(ctx, []byte(key), genkenInfo)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return key, nil
}

func (eh *KeygenRequestAddedEventHandler) reportGeneratedKey(ctx context.Context, groupId [32]byte, myIndex *big.Int, genPubKey []byte) (err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return errors.WithStack(err)
	}

	var tx *types.Transaction

	err = backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		var err error
		tx, err = transactor.ReportGeneratedKey(eh.Signer, groupId, myIndex, genPubKey)
		if err != nil {
			return errors.Wrapf(err, "failed to report genereated public key. GenPubKey: %v, PartiIndex: %v", bytes.BytesToHex(genPubKey), myIndex)
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

	return
}

func (eh *KeygenRequestAddedEventHandler) publishStoredEvent(ctx context.Context, key string, genPubKeyInfo *GeneratedPubKeyInfo, parentEvtObj *dispatcher.EventObject) {
	val := events.GeneratedPubKeyInfo(*genPubKeyInfo)
	newEvt := events.GeneratedPubKeyInfoStoredEvent{
		Key: key,
		Val: val,
	}

	eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(parentEvtObj, "KeygenRequestAddedEventHandler", &newEvt, parentEvtObj.Context))
}

func (eh *KeygenRequestAddedEventHandler) publishReportedEvent(ctx context.Context, evt *events.ReportedGenPubKeyEvent, parentEvtObj *dispatcher.EventObject) {
	eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(parentEvtObj, "KeygenRequestAddedEventHandler", evt, parentEvtObj.Context))
}
