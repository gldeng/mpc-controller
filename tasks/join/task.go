package join

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/taskcontext"
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
	//RemainingAttempts int
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
		//RemainingAttempts: 2,
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
			var errCreateTransactor *taskcontext.ErrTypTransactorCreate
			var errExecutionReverted *taskcontext.ErrTypTxReverted
			if errors.As(err, &errCreateTransactor) || errors.As(err, &errExecutionReverted) {
				ctx.GetLogger().Error(ErrMsgJoinRequest, []logger.Field{{"reqHash", fmt.Sprintf("%x", t.RequestHash)},
					{"error", err.Error()}}...)
				return nil, t.failIfErrorf(err, ErrMsgJoinRequest)
			}
			ctx.GetLogger().Debug(ErrMsgJoinRequest, []logger.Field{{"reqHash", fmt.Sprintf("%x", t.RequestHash)},
				{"error", err.Error()}}...)
			return nil, errors.Wrap(err, ErrMsgJoinRequest)
		}
		t.TxHash = *txHash
		t.Status = StatusTxSent
	case StatusTxSent:
		_, err := ctx.CheckEthTx(t.TxHash)
		if err != nil {
			ctx.GetLogger().Error(ErrMsgCheckTxStatus, []logger.Field{{"tx", t.TxHash.Hex()},
				{"reqHash", fmt.Sprintf("%x", t.RequestHash)},
				{"group", fmt.Sprintf("%x", t.group.GroupId)},
				{"error", err.Error()}}...)
			// TODO: Figure out why sometimes join mysteriously fail and replace this workaround
			//		// https://github.com/AvaLido/mpc-controller/issues/98
			if errors.Is(err, taskcontext.ErrTxAborted) {
				ctx.GetLogger().Debug("tx aborted", []logger.Field{{"tx", t.TxHash.Hex()},
					{"reqHash", fmt.Sprintf("%x", t.RequestHash)},
					{"group", fmt.Sprintf("%x", t.group.GroupId)},
					{"error", err.Error()}}...)
				t.Status = StatusInit
			}
			return nil, errors.Wrapf(err, ErrMsgCheckTxStatus)
		}

		t.Status = StatusDone
		ctx.GetLogger().Debug("joined request", []logger.Field{{"partiId", fmt.Sprintf("%x", t.group.ParticipantID())},
			{"reqHash", fmt.Sprintf("%x", t.RequestHash)},
			{"group", fmt.Sprintf("%x", t.group.GroupId)}}...)
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
