package ethlog

import (
	"context"
	"fmt"

	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
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
	groupId := fmt.Sprintf("%x", h.Event.GroupId)
	if h.Event.PublicKey != hash256.FromBytes(myPubKey) {
		ctx.GetLogger().Debug("group not for me", []logger.Field{{"group", groupId}}...)
		h.Done = true
		return nil, nil
	}

	// TODO: Add all_groups, i.e. an array containing all historical groups
	err := h.saveGroup(ctx)
	if err != nil {
		return nil, h.failIfErrorf(err, "%v for %x", ErrMsgFailedToSaveGroup, myPubKey)
	}

	h.Done = true
	ctx.GetLogger().Debug("saved group", []logger.Field{{"group", groupId}, {"myPubKey", fmt.Sprintf("%x", myPubKey)}}...)
	return nil, nil
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

	ctx.GetLogger().Debug("saving group", []logger.Field{{"group", fmt.Sprintf("%x", h.Event.GroupId)}}...)
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
	retrieved, err := ctx.GetDb().Get(context.Background(), key)
	ctx.GetLogger().Debug("retrieved group", []logger.Field{{"group", fmt.Sprintf("%x", retrieved)}}...)
	return nil
}

func (h *ParticipantAddedHandler) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}
