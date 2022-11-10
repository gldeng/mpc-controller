package ethlog

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*ParticipantAddedHandler)(nil)
)

type ParticipantAddedHandler struct {
	Event  contract.MpcManagerParticipantAdded
	Done   bool
	Failed bool
}

func (h *ParticipantAddedHandler) GetId() string {
	return fmt.Sprintf("Event(%v, %v)", h.Event.Raw.TxHash, h.Event.Raw.TxIndex)
}

func (h *ParticipantAddedHandler) FailedPermanently() bool {
	return h.Failed
}

func NewParticipantAddedHandler(event contract.MpcManagerParticipantAdded) *ParticipantAddedHandler {
	return &ParticipantAddedHandler{Event: event}
}

func (h *ParticipantAddedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	myPubKey, _ := ctx.GetMyPublicKey()
	if h.Event.PublicKey != hash256.FromBytes(myPubKey) {
		ctx.GetLogger().Debugf("Group %v not for me", bytes.Bytes32ToHex(h.Event.GroupId)) // TODO: %x
		h.Failed = true                                                                    // TODO: this expression is ambiguous
		return nil, nil
	}

	// TODO: Add all_groups, i.e. an array containing all historical groups
	err := h.saveGroup(ctx)
	if err != nil {
		ctx.GetLogger().Errorf("%v for %x, error:%+v", ErrMsgFailedToSaveGroup, myPubKey, err)
	} else {
		ctx.GetLogger().Debugf("Saved group %x for %x", h.Event.GroupId, myPubKey)
	}
	return nil, h.failIfError(err, fmt.Sprintf("%v for %x", ErrMsgFailedToSaveGroup, myPubKey))
}

func (h *ParticipantAddedHandler) IsDone() bool {
	return h.Done
}

func (h *ParticipantAddedHandler) IsSequential() bool {
	return true
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
