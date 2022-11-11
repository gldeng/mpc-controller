package stake

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
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
	reqHash types.RequestHash

	Join   *join.Join
	Failed bool
	Done   bool
}

func (t *RequestAddedHandler) GetId() string {
	hash, _ := t.Request.Hash()
	return fmt.Sprintf("JoinStake(%x)", hash)
}

func (t *RequestAddedHandler) IsDone() bool {
	return t.Done
}

func (t *RequestAddedHandler) IsSequential() bool {
	return t.Join != nil && t.Join.IsSequential()
}

func (t *RequestAddedHandler) FailedPermanently() bool {
	return t.Failed
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

	hash, err := request.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get request hash")
	}

	h := &RequestAddedHandler{
		Event:   event,
		Request: request,
		reqHash: hash,
		Join:    nil,
		Failed:  false,
	}

	return h, nil
}

func (t *RequestAddedHandler) Next(ctx core.TaskContext) ([]core.Task, error) {
	err := t.saveRequest(ctx)
	if err != nil {
		ctx.GetLogger().Errorf("Failed to save request %x, error:%+v", t.reqHash, err)
		return nil, t.failIfError(err, fmt.Sprintf("failed to save request %x", t.reqHash))
	}

	if t.Join == nil {
		join := join.NewJoin(t.reqHash)
		if join == nil {
			t.Failed = true
			return nil, errors.Errorf("invalid sub joining task created for request %+x", t.reqHash)
		}
		t.Join = join
	}

	next, err := t.Join.Next(ctx)
	if err != nil {
		ctx.GetLogger().Debugf("subtask got an error to join request %x, error:%+v", t.reqHash, err)
	}

	if t.Join.FailedPermanently() {
		ctx.GetLogger().Debugf("subtask failed to join request %x permanently, error:%+v", t.reqHash, err)
		return next, t.failIfError(err, fmt.Sprintf("subtask failed to join request %x permanently", t.reqHash))
	}

	if t.Join.IsDone() {
		t.Done = true
		ctx.GetLogger().Debugf("Joined request %x", t.reqHash)
		return next, nil
	}

	ctx.GetLogger().Debugf("Sub joining task not done, requestHash:%x", t.reqHash)
	return next, nil
}

func (t *RequestAddedHandler) saveRequest(ctx core.TaskContext) error {

	rBytes, err := t.Request.Encode()
	if err != nil {
		return err
	}

	key := []byte("request/")
	key = append(key, t.reqHash[:]...)
	return ctx.GetDb().Set(context.Background(), key, rBytes)
}

func (t *RequestAddedHandler) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, msg)
}
