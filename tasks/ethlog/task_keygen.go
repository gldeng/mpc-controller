package ethlog

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*KeyGeneratedHandler)(nil)
)

type KeyGeneratedHandler struct {
	Event  contract.MpcManagerKeyGenerated
	group  *types.Group
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
	group, err := ctx.LoadGroup(h.Event.GroupId)
	if err != nil {
		return nil, h.failIfErrorf(err, "%s %x", ErrMsgFailedToLoadGroup, h.Event.GroupId)
	}

	h.group = group

	err = h.saveKey(ctx)
	if err != nil {
		return nil, h.failIfErrorf(err, "failed to save generated public key %x for group %x", h.Event.PublicKey, group.GroupId)
	}

	ctx.GetLogger().Debugf("saved generated public key %x for group %x", h.Event.PublicKey, group.GroupId)
	h.Done = true
	return nil, nil
}

func (h *KeyGeneratedHandler) IsDone() bool {
	return h.Done
}

func (h *KeyGeneratedHandler) IsSequential() bool {
	return true
}

func (h *KeyGeneratedHandler) saveKey(ctx core.TaskContext) error {
	pubKey := &types.MpcPublicKey{
		GroupId:            h.group.GroupId,
		GenPubKey:          h.Event.PublicKey,
		ParticipantPubKeys: h.group.MemberPublicKeys,
	}

	pubKeyBytes, err := pubKey.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode generated public key")
	}
	key := []byte("latestPubKey")
	return ctx.GetDb().Set(context.Background(), key, pubKeyBytes)
}

func (h *KeyGeneratedHandler) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}
