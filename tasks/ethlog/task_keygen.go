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
	_ core.Task = (*KeyGeneratedHandler)(nil)
)

type KeyGeneratedHandler struct {
	Event   contract.MpcManagerKeyGenerated
	Done    bool
	Failed  bool
	Dropped bool
}

func (h *KeyGeneratedHandler) GetId() string {
	return fmt.Sprintf("Event(%v, %v)", h.Event.Raw.TxHash, h.Event.Raw.TxIndex)
}

func (h *KeyGeneratedHandler) FailedPermanently() bool {
	return h.Dropped
}

func NewKeyGeneratedHandler(event contract.MpcManagerKeyGenerated) *KeyGeneratedHandler {
	return &KeyGeneratedHandler{Event: event}
}

func (h *KeyGeneratedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	if len(h.Event.Raw.Topics) < 2 {
		// Do nothing, invalid event
		return nil, nil
	}
	groupId := ctx.GetParticipantID().GroupId()
	if h.Event.Raw.Topics[1] != common.BytesToHash(crypto.Keccak256(groupId[:])) {
		// Not for me
		return nil, nil
	}

	err := h.saveKey(ctx)
	return nil, h.failIfError(err, "failed to save key")
}

func (h *KeyGeneratedHandler) IsDone() bool {
	return h.Done
}

func (h *KeyGeneratedHandler) RequiresNonce() bool {
	return false
}

func (h *KeyGeneratedHandler) saveKey(ctx core.TaskContext) error {

	group, err := h.retrieveGroup(ctx)
	if err != nil {
		h.Dropped = true
		return errors.Wrap(err, "failed to get group")
	}

	pubKey := &types.MpcPublicKey{
		GroupId:            group.GroupId,
		GenPubKey:          h.Event.PublicKey,
		ParticipantPubKeys: group.MemberPublicKeys,
	}

	if err != nil {
		return err
	}
	pubKeyBytes, err := pubKey.Encode()
	if err != nil {
		return err
	}
	key := []byte("latestPubKey")
	return ctx.GetDb().Set(context.Background(), key, pubKeyBytes)
}

func (h *KeyGeneratedHandler) retrieveGroup(ctx core.TaskContext) (*types.Group, error) {
	groupId := ctx.GetParticipantID().GroupId()
	groupKey := []byte("group/")
	groupKey = append(groupKey, groupId[:]...)
	groupBytes, err := ctx.GetDb().Get(context.Background(), groupKey)
	if err != nil {
		return nil, err
	}
	group := &types.Group{}
	err = group.Decode(groupBytes)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (h *KeyGeneratedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, msg)
}
