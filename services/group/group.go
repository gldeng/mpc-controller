package group

import (
	"context"
	ctlPk "github.com/avalido/mpc-controller"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// ---------------------------------------------------------------------------------------------------------------------
// Group implementation

var _ ctlPk.MpcControllerService = (*Group)(nil)

// Group instance watches ParticipantAdded event emitted from MpcManager contract,
// which will result in local persistence of corresponding ParticipantInfo and GroupInfo datum.
type Group struct {
	Log           logger.Logger
	PubKeyStr     string
	PubKeyBytes   []byte
	PubKeyHashStr string

	ctlPk.CallerGetGroup          // MpcManager contract caller
	ctlPk.WatcherParticipantAdded // MpcManager filter

	ctlPk.StorerStoreGroupInfo
	ctlPk.StorerStoreParticipantInfo

	participantAddedEvt chan *contract.MpcManagerParticipantAdded
}

func (p *Group) Start(ctx context.Context) error {
	// Watch ParticipantAdded event
	go func() {
		err := p.watchParticipantAdded(ctx)
		p.Log.ErrorOnError(err, "Got an error to watch ParticipantAdded event")
	}()

	// Store participant added
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-p.participantAddedEvt:
			err := p.storeParticipantAdded(evt)
			p.Log.ErrorOnError(err, "Failed to process ParticipantAdded event")
		}
	}
}

func (p *Group) watchParticipantAdded(ctx context.Context) error {
	// Subscribe ParticipantAdded event
	pubKeys := [][]byte{
		p.PubKeyBytes,
	}
	sink, err := p.WatchParticipantAdded(pubKeys)
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch ParticipantAdded event
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-sink:
			p.Log.WarnOnNotOk(ok, "Retrieve nothing from event channel of ParticipantAdded")
			if ok {
				p.participantAddedEvt <- evt
			}
		}
	}
}

func (p *Group) storeParticipantAdded(evt *contract.MpcManagerParticipantAdded) error {
	// Store participant
	groupId := common.Bytes2Hex(evt.GroupId[:])
	pt := storage.ParticipantInfo{
		PubKeyHashHex: p.PubKeyHashStr,
		PubKeyHex:     p.PubKeyStr,
		GroupIdHex:    groupId,
		Index:         evt.Index.Uint64(),
	}
	err := p.StoreParticipantInfo(&pt)
	if err != nil {
		return errors.WithStack(err)
	}
	p.Log.Debug("Stored a participant", logger.Field{"participant", pt})

	// Store group
	pubKeyBytes, threshold, err := p.GetGroup(evt.GroupId)
	if err != nil {
		return errors.WithStack(err)
	}

	var pubKeys []string
	for _, k := range pubKeyBytes {
		pk := common.Bytes2Hex(k)
		pubKeys = append(pubKeys, pk)
	}

	t := threshold.Uint64()

	g := storage.GroupInfo{
		GroupIdHex:     groupId,
		PartPubKeyHexs: pubKeys,
		Threshold:      t,
	}
	err = p.StoreGroupInfo(&g)
	if err != nil {
		return errors.WithStack(err)
	}
	p.Log.Debug("Stored a group", logger.Field{"group", p})
	return nil
}
