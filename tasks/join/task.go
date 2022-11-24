package join

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
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

	group *types.Group

	TxHash    common.Hash
	Failed    bool
	StartTime time.Time
}

func (t *Join) GetId() string {
	return fmt.Sprintf("Join(%x)", t.RequestHash)
}

func (t *Join) FailedPermanently() bool {
	return t.Failed
}

func NewJoin(requestHash [32]byte) *Join {
	return &Join{
		RequestHash: requestHash,
		Status:      StatusInit,
		group:       nil,
		TxHash:      common.Hash{},
		Failed:      false,
		StartTime:   time.Now(),
	}
}

func (t *Join) Next(ctx core.TaskContext) ([]core.Task, error) {
	if t.group == nil {
		group, err := ctx.LoadGroupByLatestMpcPubKey() // TODO: should we always use the latest one?
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to load group for joining request %x", t.RequestHash)
		}
		t.group = group
	}

	timeOut := 30 * time.Minute
	interval := 2 * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop()
	var next []core.Task
	var err error

	for {
		select {
		case <-timer.C:
			next, err = t.run(ctx)
			if t.Status == StatusDone || t.Failed {
				return next, errors.Wrap(err, "failed to export from C-Chain")
			}
			if time.Now().Sub(t.StartTime) >= timeOut {
				return nil, errors.New(ErrMsgTimedOut)
			}

			timer.Reset(interval)
		}
	}

	return t.run(ctx)
}

func (t *Join) IsDone() bool {
	return t.Status == StatusDone
}

func (t *Join) IsSequential() bool {
	return true
}

func (t *Join) run(ctx core.TaskContext) ([]core.Task, error) {
	switch t.Status {
	case StatusInit:
		txHash, err := ctx.JoinRequest(ctx.GetMyTransactSigner(), t.group.ParticipantID(), t.RequestHash)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to join request %x", t.RequestHash)
		}
		t.TxHash = *txHash
		t.Status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckEthTx(t.TxHash)
		ctx.GetLogger().Debugf("id %v Join Status is %v", t.GetId(), status)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to check status for tx %x", t.TxHash)
		}
		switch status {
		case core.TxStatusUnknown:
			return nil, t.failIfErrorf(errors.Errorf("unkonw tx status (%v:%x) of joining request %x", status, t.TxHash, t.RequestHash), "")
		case core.TxStatusAborted:
			t.Status = StatusInit // TODO: avoid endless repeating joining?
			return nil, t.failIfErrorf(errors.Errorf("joining request %x tx %x aborted for group %x", t.RequestHash, t.TxHash, t.group.GroupId), "")
		case core.TxStatusCommitted:
			t.Status = StatusDone
			ctx.GetLogger().Debugf("Joined request. participantId:%x requestHash:%x group:%x", t.group.ParticipantID(), t.RequestHash, t.group.GroupId)
		}
	}
	return nil, nil
}

func (t *Join) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}
