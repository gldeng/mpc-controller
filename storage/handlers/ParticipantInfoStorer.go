package handlers

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
)

type Storer interface {
	storage.StorerStoreParticipantInfo
}

// Trigger event: *contract.MpcManagerParticipantAdded
// Emit event: *events.ParticipantInfoStoredEvent

type ParticipantInfoStorer struct {
	Logger logger.Logger

	Publisher dispatcher.Publisher
	Storer    Storer

	PubKeyHex     string
	PubKeyHashHex string
}

func (p *ParticipantInfoStorer) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerParticipantAdded:
		p.storeParticipantInfo(evtObj.Context, evt, evtObj)
	}
}

func (p *ParticipantInfoStorer) storeParticipantInfo(ctx context.Context, evt *contract.MpcManagerParticipantAdded, evtObj *dispatcher.EventObject) {
	pt := storage.ParticipantInfo{
		PubKeyHashHex: p.PubKeyHashHex,
		PubKeyHex:     p.PubKeyHex,
		GroupIdHex:    bytes.Bytes32ToHex(evt.GroupId),
		Index:         evt.Index.Uint64(),
	}
	err := p.Storer.StoreParticipantInfo(ctx, &pt)
	p.Logger.ErrorOnError(err, "Fail to store participantInfo",
		[]logger.Field{{"error", err}, {"participantInfo", &pt}}...)

	newEvt := events.ParticipantInfoStoredEvent{
		PubKeyHashHex: pt.PubKeyHashHex,
		PubKeyHex:     pt.PubKeyHex,
		GroupIdHex:    pt.GroupIdHex,
		Index:         pt.Index,
	}

	p.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "ParticipantInfoStorer", &newEvt, evtObj.Context))
}
