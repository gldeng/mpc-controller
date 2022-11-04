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
	Event  contract.MpcManagerKeyGenerated
	Done   bool
	Failed bool
}

func (h *KeyGeneratedHandler) GetId() string {
	return fmt.Sprintf("Event(%v, %v)", h.Event.Raw.TxHash, h.Event.Raw.TxIndex)
}

func (h *KeyGeneratedHandler) FailedPermanently() bool {
	return h.Failed
}

func NewKeyGeneratedHandler(event contract.MpcManagerKeyGenerated) *KeyGeneratedHandler {
	return &KeyGeneratedHandler{Event: event}
}

func (h *KeyGeneratedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	groupId := ctx.GetParticipantID().GroupId()
	if h.Event.Raw.Topics[1] != common.BytesToHash(crypto.Keccak256(groupId[:])) {
		h.Done = true // TODO: this expression is ambiguous
		return nil, nil
	}

	err := h.saveKey(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("failed to save generated public key %x", h.Event.PublicKey)
		ctx.GetLogger().ErrorOnError(err, errMsg)
		return nil, h.failIfError(err, errMsg)
	}

	ctx.GetLogger().Debug(fmt.Sprintf("saved generated public key %v", h.Event.PublicKey))
	h.Done = true
	return nil, nil
}

func (h *KeyGeneratedHandler) IsDone() bool {
	return h.Done
}

func (h *KeyGeneratedHandler) RequiresNonce() bool {
	return false
}

func (h *KeyGeneratedHandler) saveKey(ctx core.TaskContext) error {
	group, err := ctx.LoadGroup(h.Event.GroupId)
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("failed to load group %x", h.Event.GroupId))
	}

	pubKey := &types.MpcPublicKey{
		GroupId:            group.GroupId,
		GenPubKey:          h.Event.PublicKey,
		ParticipantPubKeys: group.MemberPublicKeys,
	}

	pubKeyBytes, err := pubKey.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode generated public key")
	}
	key := []byte("latestPubKey")
	return ctx.GetDb().Set(context.Background(), key, pubKeyBytes)
}

//func (h *KeyGeneratedHandler) retrieveGroup(ctx core.TaskContext) (*types.Group, error) {
//	groupId := ctx.GetParticipantID().GroupId()
//	groupKey := []byte("group/")
//	groupKey = append(groupKey, groupId[:]...)
//	groupBytes, err := ctx.GetDb().Get(context.Background(), groupKey)
//	if err != nil {
//		return nil, err
//	}
//	group := &types.Group{}
//	err = group.Decode(groupBytes)
//	if err != nil {
//		return nil, err
//	}
//	return group, nil
//}

func (h *KeyGeneratedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, msg)
}
