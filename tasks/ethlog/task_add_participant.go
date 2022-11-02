package ethlog

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*ParticipantAddedHandler)(nil)
)

type ParticipantAddedHandler struct {
	Event  contract.MpcManagerParticipantAdded
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
	if len(h.Event.Raw.Topics) < 2 {
		// Do nothing, invalid event
		return nil, nil
	}
	pubKey, _ := ctx.GetMyPublicKey()
	if h.Event.Raw.Topics[1] != common.BytesToHash(crypto.Keccak256(pubKey)) {
		// Not for me
		return nil, nil
	}
	// TODO: Add all_groups, i.e. an array containing all historical groups
	err := h.saveGroup(ctx)
	return nil, h.failIfError(err, "failed to save group")
}

func (h *ParticipantAddedHandler) IsDone() bool {
	return true
}

func (h *ParticipantAddedHandler) RequiresNonce() bool {
	return false
}

func (h *ParticipantAddedHandler) saveGroup(ctx core.TaskContext) error {
	group := types.Group{
		GroupId: h.Event.GroupId,
		Index:   h.Event.Index,
	}
	key := []byte("group/")
	key = append(key, group.GroupId[:]...)
	groupBytes, err := group.Encode()
	if err != nil {
		return err
	}
	err = ctx.GetDb().Set(context.Background(), key, groupBytes)
	if err != nil {
		return err
	}
	return err
}

func (h *ParticipantAddedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, msg)
}
