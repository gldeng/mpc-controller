package ethlog

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/stake"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*RequestStartedHandler)(nil)
)

type RequestStartedHandler struct {
	Event  contract.MpcManagerRequestStarted
	Failed bool
	Done   bool
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
	if err != nil {
		return nil, h.failIfErrorf(err, "failed to check index")
	}
	if !p {
		h.Done = true
		ctx.GetLogger().Debugf("not participate %x", h.Event.RequestHash)
		return nil, nil
	}
	hash := (types.RequestHash)(h.Event.RequestHash)
	data, err := h.retrieveRequest(ctx)
	if err != nil {
		return nil, h.failIfErrorf(err, "failed to retrieve request")
	}

	var tasks []core.Task
	switch hash.TaskType() {
	case types.TaskTypStake:
		r := new(stake.Request)
		err := r.Decode(data)
		if err != nil {
			return nil, h.failIfErrorf(err, "failed decode request")
		}
		ctx.GetLogger().Debug("stake request", logger.Field{
			Key:   "request",
			Value: r,
		})
		pubKeyKey := []byte("latestPubKey") // TODO: Always use latest public key?
		quorum, err := h.getQuorumInfo(ctx, pubKeyKey)
		if err != nil {
			return nil, h.failIfErrorf(err, "failed to get quorum info")
		}
		task, err := stake.NewInitialStake(r, *quorum)
		if err != nil {
			return nil, h.failIfErrorf(err, "failed to create stake task")
		}
		tasks = append(tasks, task)
	case types.TaskTypRecover:
		// TODO: Add logics here
	}

	h.Done = true
	return tasks, nil
}

func (h *RequestStartedHandler) IsDone() bool {
	return h.Done
}

func (h *RequestStartedHandler) IsSequential() bool {
	return false
}

func (h *RequestStartedHandler) participating(ctx core.TaskContext) (bool, error) {
	group, err := ctx.LoadGroupByLatestMpcPubKey()
	if err != nil {
		return false, errors.Wrap(err, "failed to load group by latest mpc public key")
	}
	partiID := group.ParticipantID()
	isParti := partiID.Joined(h.Event.ParticipantIndices)
	return isParti, nil
}

func (h *RequestStartedHandler) retrieveRequest(ctx core.TaskContext) ([]byte, error) {
	key := []byte("request/")
	key = append(key, h.Event.RequestHash[:]...)
	return ctx.GetDb().Get(context.Background(), key)
}

func (h *RequestStartedHandler) retrieveKey(ctx core.TaskContext, key []byte) (*types.MpcPublicKey, error) {
	bytes, err := ctx.GetDb().Get(context.Background(), key)
	if err != nil {
		return nil, h.failIfErrorf(err, "failed to load mpc key info")
	}
	keyInfo := new(types.MpcPublicKey)
	err = keyInfo.Decode(bytes)
	if err != nil {
		return nil, h.failIfErrorf(err, "failed to decode mpc key info")
	}
	return keyInfo, nil
}

func (h *RequestStartedHandler) getQuorumInfo(ctx core.TaskContext, key []byte) (*types.QuorumInfo, error) {
	keyInfo, err := h.retrieveKey(ctx, key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve public key to %x", key)
	}
	indices := ((*types.Indices)(h.Event.ParticipantIndices)).Indices()
	pks := make([]types.PubKey, len(indices))
	for i, index := range indices {
		k := keyInfo.ParticipantPubKeys[index-1]
		if k == nil {
			return nil, errors.New("can't find key")
		}
		pks[i] = k
	}
	return &types.QuorumInfo{
		ParticipantPubKeys: pks,
		PubKey:             keyInfo.GenPubKey,
	}, nil
}

func (h *RequestStartedHandler) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	h.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}
