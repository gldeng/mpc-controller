package stake

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/tasks/join"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*RequestAddedHandler)(nil)
)

type RequestAddedHandler struct {
	Event   contract.MpcManagerStakeRequestAdded
	Request Request

	Join   *join.Join
	Failed bool
}

func (t *RequestAddedHandler) GetId() string {
	hash, _ := t.Request.Hash()
	return fmt.Sprintf("JoinStake(%x)", hash)
}

func (t *RequestAddedHandler) IsDone() bool {
	return t.Join != nil && t.Join.IsDone()
}

func (t *RequestAddedHandler) IsSequential() bool {
	return t.Join != nil && t.Join.IsSequential()
}

func (t *RequestAddedHandler) FailedPermanently() bool {
	return t.Join != nil && t.Join.IsSequential()
}

func NewStakeRequestAddedHandler(event contract.MpcManagerStakeRequestAdded) (*RequestAddedHandler, error) {
	request := Request{
		ReqNo:     event.RequestNumber.Uint64(),
		TxHash:    common.Hash{},
		PubKey:    event.PublicKey.Bytes(),
		NodeID:    event.NodeID,
		Amount:    event.Amount.String(),
		StartTime: event.StartTime.Uint64(),
		EndTime:   event.EndTime.Uint64(),
	}

	return &RequestAddedHandler{
		Event:   event,
		Request: request,
		Join:    nil,
		Failed:  false,
	}, nil
}

func (t *RequestAddedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	err := t.saveRequest(ctx)
	if err != nil {
		return nil, t.failIfError(err, "failed to save request")
	}
	hash, err := t.Request.Hash()
	if err != nil {
		return nil, t.failIfError(err, "failed to get hash")
	}
	t.Join = join.NewJoin(hash)
	if t.Join != nil {
		next, err := t.Join.Next(ctx)
		return next, t.failIfError(err, "failed to save request")
	}

	return nil, nil
}

func (t *RequestAddedHandler) saveRequest(ctx core.TaskContext) error {

	rBytes, err := t.Request.Encode()
	if err != nil {
		return err
	}

	hash, err := t.Request.Hash()
	if err != nil {
		return err
	}

	key := []byte("request/")
	key = append(key, hash[:]...)
	return ctx.GetDb().Set(context.Background(), key, rBytes)
}

func (t *RequestAddedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, msg)
}
