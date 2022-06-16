package participant

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
)

// Trigger event: *contract.MpcManagerParticipantAdded
// Emit event: *events.ParticipantInfoStoredEvent

type ParticipantInfoStorer struct {
	Logger logger.Logger

	Publisher dispatcher.Publisher
	Storer    storage.MarshalSetter

	MyPubKeyHex     string
	MyPubKeyHashHex string
}

func (p *ParticipantInfoStorer) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerParticipantAdded:
		key, pt, err := p.storeParticipantInfo(evtObj.Context, evt)
		if err != nil {
			p.Logger.Error("Fail to store participantInfo", []logger.Field{{"error", err}, {"participantInfo", &pt}}...)
			break
		}
		p.publishStoredEvent(evtObj.Context, key, pt, evtObj)
	}
}

func (p *ParticipantInfoStorer) storeParticipantInfo(ctx context.Context, evt *contract.MpcManagerParticipantAdded) (string, *ParticipantInfo, error) {
	pt := ParticipantInfo{
		PubKeyHashHex: p.MyPubKeyHashHex,
		PubKeyHex:     p.MyPubKeyHex,
		GroupIdHex:    bytes.Bytes32ToHex(evt.GroupId),
		Index:         evt.Index.Uint64(),
	}

	key := prefixParticipantInfo + "-" + pt.PubKeyHashHex + "-" + pt.GroupIdHex
	err := p.Storer.Set(ctx, []byte(key), p)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}

	return key, &pt, nil
}

func (p *ParticipantInfoStorer) publishStoredEvent(ctx context.Context, key string, pt *ParticipantInfo, parentEvtObj *dispatcher.EventObject) {
	val := events.ParticipantInfo(*pt)
	newEvt := events.ParticipantInfoStoredEvent{
		Key: key,
		Val: val,
	}

	p.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(parentEvtObj, "ParticipantInfoStorer", &newEvt, parentEvtObj.Context))
}
