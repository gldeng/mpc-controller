package keygen

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/pkg/errors"
)

// Accept event: *events.GroupInfoStoredEvent
// Accept event: *contract.MpcManagerKeygenRequestAdded

// Emit event: *events.GeneratedPubKeyInfoStoredEvent

type KeygenRequestAddedEventHandler struct {
	Logger logger.Logger

	KeygenDoner core.KeygenDoner
	Storer      storage.MarshalSetter

	Publisher dispatcher.Publisher

	groupInfoMap map[string]events.GroupInfo
}

// Pre-condition: *contract.MpcManagerKeygenRequestAdded must happen after *event.GroupInfoStoredEvent

func (eh *KeygenRequestAddedEventHandler) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.GroupInfoStoredEvent:
		if len(eh.groupInfoMap) == 0 {
			eh.groupInfoMap = make(map[string]events.GroupInfo)
		}
		eh.groupInfoMap[evt.Key] = evt.Val
	case *contract.MpcManagerKeygenRequestAdded:
		err := eh.do(evtObj.Context, evt, evtObj)
		eh.Logger.ErrorOnError(err, "Failed to deal with MpcManagerKeygenRequestAdded event.", []logger.Field{{"error", err}}...)
	}
}

func (eh *KeygenRequestAddedEventHandler) do(ctx context.Context, req *contract.MpcManagerKeygenRequestAdded, evtObj *dispatcher.EventObject) error {
	reqId := req.Raw.TxHash.Hex()

	groupIdHex := bytes.Bytes32ToHex(req.GroupId)
	group := eh.groupInfoMap[groupIdHex]

	keyGenReq := &core.KeygenRequest{
		RequestId:       reqId,
		ParticipantKeys: group.PartPubKeyHexs,
		Threshold:       group.Threshold,
	}

	genPubKeyHex, err := eh.keygen(ctx, keyGenReq)
	if err != nil {
		return errors.WithStack(err)
	}

	genPubKeyHash := hash256.FromHex(genPubKeyHex)

	genPubKeyInfo := GeneratedPubKeyInfo{
		GenPubKeyHashHex: genPubKeyHash.Hex(),
		GenPubKeyHex:     genPubKeyHex,
		GroupIdHex:       groupIdHex,
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

	err := eh.Storer.Set(ctx, []byte(key), genkenInfo)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return key, nil
}

func (eh *KeygenRequestAddedEventHandler) publishStoredEvent(ctx context.Context, key string, genPubKeyInfo *GeneratedPubKeyInfo, parentEvtObj *dispatcher.EventObject) {
	val := events.GeneratedPubKeyInfo(*genPubKeyInfo)
	newEvt := events.GeneratedPubKeyInfoStoredEvent{
		Key: key,
		Val: val,
	}

	eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(parentEvtObj, "GroupInfoStorer", &newEvt, parentEvtObj.Context))
}
