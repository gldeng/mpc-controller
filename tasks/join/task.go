package join

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*Join)(nil)
)

type Join struct {
	RequestHash [32]byte
	Status      Status

	group *types.Group

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
	if t.group == nil {
		group, err := ctx.LoadGroupByLatestMpcPubKey() // TODO: should we always use the latest one?
		if err != nil {
			ctx.GetLogger().Errorf("failed to load group for joining request %x, error:%+v", t.RequestHash, err)
			return nil, t.failIfError(err, fmt.Sprintf("failed to load group for joining request %x", t.RequestHash))
		}
		t.group = group
	}

	//interval := 100 * time.Millisecond
	//timer := time.NewTimer(interval)
	//for {
	//	select {
	//	case <-timer.C:
	//		next, err := t.run(ctx)
	//		if err != nil || t.Status == StatusDone {
	//			return next, err
	//		} else {
	//			timer.Reset(interval)
	//		}
	//	}
	//}

	//return nil, nil //
	return t.run(ctx)
}

func (t *Join) IsDone() bool {
	return t.Status == StatusDone
}

func (t *Join) IsSequential() bool {
	return false
}

func (t *Join) run(ctx core.TaskContext) ([]core.Task, error) {
	switch t.Status {
	case StatusInit:
		txHash, err := ctx.JoinRequest(ctx.GetMyTransactSigner(), t.group.ParticipantID(), t.RequestHash)
		if err != nil {
			ctx.GetLogger().Debugf("Failed to join request. participantId:%x requestHash:%x group:%x, error:%+v", t.group.ParticipantID(), t.RequestHash, t.group.GroupId, err)
			return nil, t.failIfError(err, fmt.Sprintf("failed to join request %x", t.RequestHash))
		}
		t.TxHash = *txHash
		t.Status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckEthTx(t.TxHash)
		ctx.GetLogger().Debugf("id %v Join Status is %v", t.GetId(), status)
		if err != nil {
			ctx.GetLogger().Errorf("Failed to check status for tx %x, error:%+v", t.TxHash, err)
			return nil, t.failIfError(err, fmt.Sprintf("failed to check status for tx %x", t.TxHash))
		}
		switch status {
		case core.TxStatusUnknown:
			ctx.GetLogger().Debugf("Unkonw tx status (%v:%x) of joing request %x", status, t.TxHash, t.RequestHash)
			return nil, t.failIfError(errors.Errorf("unkonw tx status (%v:%x) of joining request %x", status, t.TxHash, t.RequestHash), "")
		case core.TxStatusAborted:
			t.Status = StatusInit // TODO: avoid endless repeating joining?
			ctx.GetLogger().Errorf("Joining request %x tx %x aborted for group %x", t.RequestHash, t.TxHash, t.group.GroupId)
			return nil, errors.Errorf("joining request %x tx %x aborted for group %x", t.RequestHash, t.TxHash, t.group.GroupId)
		case core.TxStatusCommitted:
			t.Status = StatusDone
			ctx.GetLogger().Debugf("Joined request. participantId:%x requestHash:%x group:%x", t.group.ParticipantID(), t.RequestHash, t.group.GroupId)
		}
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
