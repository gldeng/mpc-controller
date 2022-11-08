package join

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"time"
)

var (
	_ core.Task = (*Join)(nil)
)

type Join struct {
	RequestHash [32]byte
	Status      Status

	TxHash common.Hash
	Failed bool
}

func (t *Join) GetId() string {
	return fmt.Sprintf("Join(%x)", t.RequestHash)
}

func (t *Join) FailedPermanently() bool {
	return t.Failed
}

func NewJoin(requestHash [32]byte) *Join {
	return &Join{RequestHash: requestHash}
}

func (t *Join) Next(ctx core.TaskContext) ([]core.Task, error) {
	interval := 100 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		select {
		case <-timer.C:
			next, err := t.run(ctx)
			if err != nil || t.Status == StatusDone {
				return next, err
			} else {
				timer.Reset(interval)
			}
		}
	}

	return nil, nil
}

func (t *Join) IsDone() bool {
	return true
}

func (t *Join) RequiresNonce() bool {
	return false
}

func (t *Join) run(ctx core.TaskContext) ([]core.Task, error) {
	switch t.Status {
	case StatusInit:
		participantId := ctx.GetParticipantID()

		txHash, err := ctx.JoinRequest(nil, participantId, t.RequestHash)
		t.TxHash = *txHash
		if err != nil {
			return nil, errors.Wrap(err, "failed to join request")
		} else {
			t.Status = StatusTxSent
		}
	case StatusTxSent:
		status, err := ctx.CheckEthTx(t.TxHash)
		ctx.GetLogger().Debugf("id %v ReportGeneratedKey Status is %v", t.GetId(), status)
		if err != nil {
			return nil, t.failIfError(err, "failed to check tx status")
		}
		if !core.IsPending(status) {
			t.Status = StatusDone
			return nil, nil
		}
		return nil, nil
	}
	return nil, nil
}

func (t *Join) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, msg)
}
