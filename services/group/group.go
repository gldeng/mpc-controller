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
	PubKeyStr     string
	PubKeyBytes   []byte
	PubKeyHashStr string

	logger.Logger
	ctlPk.CallerGetGroup          // MpcManager contract caller
	ctlPk.WatcherParticipantAdded // MpcManager filter

	ctlPk.StorerStoreGroupInfo
	ctlPk.StorerStoreParticipantInfo

	participantAddedEvt chan *contract.MpcManagerParticipantAdded
}

func (p *Group) Start(ctx context.Context) error {
	// Assign unexported fields
	p.participantAddedEvt = make(chan *contract.MpcManagerParticipantAdded)

	// Watch ParticipantAdded event
	go func() {
		err := p.watchParticipantAdded(ctx)
		p.ErrorOnError(err, "Got an error to watch ParticipantAdded event", logger.Field{"error", err})
	}()

	// Store participant added
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-p.participantAddedEvt:
			err := p.onParticipantAdded(ctx, evt)
			p.ErrorOnError(err, "Failed to process ParticipantAdded event", logger.Field{"error", err})
		}
	}
}

func (p *Group) watchParticipantAdded(ctx context.Context) error {
	// Subscribe ParticipantAdded event
	pubKeys := [][]byte{
		p.PubKeyBytes,
	}
	sink, err := p.WatchParticipantAdded(ctx, pubKeys)
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch ParticipantAdded event
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-sink:
			p.WarnOnNotOk(ok, "Retrieve nothing from event channel of ParticipantAdded")
			if ok {
				p.participantAddedEvt <- evt
			}
		}
	}
}

func (p *Group) onParticipantAdded(ctx context.Context, evt *contract.MpcManagerParticipantAdded) error {
	// Store participant
	groupId := common.Bytes2Hex(evt.GroupId[:])
	pt := storage.ParticipantInfo{
		PubKeyHashHex: p.PubKeyHashStr,
		PubKeyHex:     p.PubKeyStr,
		GroupIdHex:    groupId,
		Index:         evt.Index.Uint64(),
	}
	err := p.StoreParticipantInfo(ctx, &pt)
	if err != nil {
		return errors.WithStack(err)
	}
	p.Debug("Stored a participant", logger.Field{"participant", pt})

	// Store group
	pubKeyBytes, threshold, err := p.GetGroup(ctx, evt.GroupId)
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
	err = p.StoreGroupInfo(ctx, &g)
	if err != nil {
		return errors.WithStack(err)
	}
	p.Debug("Stored a group", logger.Field{"group", g})
	return nil
}
