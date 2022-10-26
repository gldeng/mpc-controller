package ethlog

import (
	"context"
	"encoding/json"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	_ core.Task    = (*RequestStartedHandler)(nil)
	_ core.Request = (*StakeRequest)(nil)
)

type RequestStartedHandler struct {
	Event contract.MpcManagerRequestStarted
}

func NewRequestStartedHandler(event contract.MpcManagerRequestStarted) *RequestStartedHandler {
	return &RequestStartedHandler{Event: event}
}

func (h *RequestStartedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	p, err := h.participating(ctx)
	if !p {
		return nil, errors.Wrap(err, "failed to check index")
	}
	hash := (storage.RequestHash)(h.Event.RequestHash)
	data, err := h.retrieveRequest(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve request")
	}
	switch hash.TaskType() {
	case storage.TaskTypStake:
		r := new(StakeRequest)
		r.Decode(data)
		// TODO: Add stake task
	case storage.TaskTypRecover:
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
	for _, index := range ((*storage.Indices)(h.Event.ParticipantIndices)).Indices() {
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

// TODO: Create StakeTask and move StakeRequest to that package
type StakeRequest struct {
	ReqNo     uint64      `json:"reqNo"`
	TxHash    common.Hash `json:"txHash"`
	NodeID    string      `json:"nodeID"`
	Amount    string      `json:"amount"`
	StartTime int64       `json:"startTime"`
	EndTime   int64       `json:"endTime"`
}

func (r *StakeRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r *StakeRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r StakeRequest) Hash() (storage.RequestHash, error) {
	data, err := r.Encode()
	if err != nil {
		return [32]byte{}, err
	}
	hash := storage.RequestHash(hash256.FromBytes(data))
	hash.SetTaskType(storage.TaskTypStake)
	return hash, nil
}
