package ethlog

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*ParticipantAddedHandler)(nil)
)

type ParticipantAddedHandler struct {
	Event   contract.MpcManagerParticipantAdded
	Done    bool
	Failed  bool // Failed task need be re-enqueued, error maybe cause by network partition, ect.
	Dropped bool // Dropped task means failed permanently and shouldn't be re-enqueued
}

func (h *ParticipantAddedHandler) GetId() string {
	return fmt.Sprintf("Event(%v, %v)", h.Event.Raw.TxHash, h.Event.Raw.TxIndex)
}

func (h *ParticipantAddedHandler) FailedPermanently() bool {
	return h.Dropped
}

func NewParticipantAddedHandler(event contract.MpcManagerParticipantAdded) *ParticipantAddedHandler {
	return &ParticipantAddedHandler{Event: event}
}

func (h *ParticipantAddedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	if len(h.Event.Raw.Topics) < 2 {
		h.Dropped = true
		return nil, errors.Errorf("invalid event topics length")
	}
	myPubKey, _ := ctx.GetMyPublicKey()
	if h.Event.PublicKey != hash256.FromBytes(myPubKey) {
		h.Dropped = true
		return nil, nil
	}

	// TODO: Add all_groups, i.e. an array containing all historical groups
	err := h.saveGroup(ctx)
	if err != nil {
		h.Dropped = true
	}
	ctx.GetLogger().DebugNilError(err, fmt.Sprintf("saved group for %x", myPubKey))
	errMsg := fmt.Sprintf("failed to save group for %x", myPubKey)
	ctx.GetLogger().ErrorOnError(err, errMsg)
	return nil, h.failIfError(err, errMsg)
}

func (h *ParticipantAddedHandler) IsDone() bool {
	return h.Done
}

func (h *ParticipantAddedHandler) RequiresNonce() bool {
	return false
}

func (h *ParticipantAddedHandler) saveGroup(ctx core.TaskContext) error {
	members, err := ctx.GetGroup(nil, h.Event.GroupId)
	if err != nil {
		return errors.Wrap(err, "failed to get group")
	}

	group := types.Group{
		GroupId:          h.Event.GroupId,
		Index:            h.Event.Index,
		MemberPublicKeys: members,
	}
	key := []byte("group/")
	key = append(key, group.GroupId[:]...)
	groupBytes, err := group.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode group")
	}
	// Note: what if participant and group has many-many relationship? In this case group shouldn't be overwritten?
	err = ctx.GetDb().Set(context.Background(), key, groupBytes)
	if err != nil {
		return errors.Wrap(err, "failed to set group")
	}
	h.Done = true
	return nil
}

func (h *ParticipantAddedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, msg)
}
