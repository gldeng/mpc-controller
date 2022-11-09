package ethlog

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/tasks/stake"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*RequestStartedHandler)(nil)
)

type RequestStartedHandler struct {
	Event  contract.MpcManagerRequestStarted
	Failed bool
}

func (h *RequestStartedHandler) GetId() string {
	return fmt.Sprintf("Event(%v, %v)", h.Event.Raw.TxHash, h.Event.Raw.TxIndex)
}

func (h *RequestStartedHandler) FailedPermanently() bool {
	return h.Failed
}

func NewRequestStartedHandler(event contract.MpcManagerRequestStarted) *RequestStartedHandler {
	return &RequestStartedHandler{Event: event}
}

func (h *RequestStartedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	p, err := h.participating(ctx)
	if !p {
		return nil, h.failIfError(err, "failed to check index")
	}
	hash := (types.RequestHash)(h.Event.RequestHash)
	data, err := h.retrieveRequest(ctx)
	if err != nil {
		return nil, h.failIfError(err, "failed to retrieve request")
	}
	switch hash.TaskType() {
	case types.TaskTypStake:
		r := new(stake.Request)
		err := r.Decode(data)
		if err != nil {
			return nil, h.failIfError(err, "failed decode request")
		}
		quorum, err := h.getQuorumInfo(ctx, r.PubKey)
		if err != nil {
			return nil, h.failIfError(err, "failed to get quorum info")
		}
		task, err := stake.NewInitialStake(r, *quorum)
		if err != nil {
			return nil, h.failIfError(err, "failed to create stake task")
		}
		return []core.Task{task}, nil
	case types.TaskTypRecover:
	// TODO: Add logics here
	default:
		return nil, nil
	}
	return nil, nil
}

func (h *RequestStartedHandler) IsDone() bool {
	return true
}

func (h *RequestStartedHandler) RequiresNonce() bool {
	return false
}

func (h *RequestStartedHandler) participating(ctx core.TaskContext) (bool, error) {
	id := ctx.GetParticipantID()
	myIndex := id.Index()
	for _, index := range ((*types.Indices)(h.Event.ParticipantIndices)).Indices() {
		if myIndex == uint64(index) {
			return true, nil
		}
	}
	return false, nil
}

func (h *RequestStartedHandler) retrieveRequest(ctx core.TaskContext) ([]byte, error) {
	key := []byte("request/")
	key = append(key, h.Event.RequestHash[:]...)
	return ctx.GetDb().Get(context.Background(), key)
}

func (h *RequestStartedHandler) retrieveKey(ctx core.TaskContext, pubKey []byte) (*types.MpcPublicKey, error) {
	key := []byte("key/")
	key = append(key, pubKey...)
	bytes, err := ctx.GetDb().Get(context.Background(), key)
	if err != nil {
		return nil, h.failIfError(err, "failed to load mpc key info")
	}
	keyInfo := new(types.MpcPublicKey)
	err = keyInfo.Decode(bytes)
	if err != nil {
		return nil, h.failIfError(err, "failed to decode mpc key info")
	}
	return keyInfo, nil
}

func (h *RequestStartedHandler) getQuorumInfo(ctx core.TaskContext, pubKey []byte) (*types.QuorumInfo, error) {
	keyInfo, err := h.retrieveKey(ctx, pubKey)
	if err != nil {
		return nil, err
	}
	indices := ((*types.Indices)(h.Event.ParticipantIndices)).Indices()
	pks := make([][]byte, len(indices))
	for i, index := range indices {
		k := keyInfo.ParticipantPubKeys[index-1]
		if k == nil {
			return nil, errors.New("can't find key")
		}
		pks[i] = k
	}
	return &types.QuorumInfo{
		ParticipantPubKeys: pks,
		PubKey:             pubKey,
	}, nil
}

func (h *RequestStartedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, msg)
}
