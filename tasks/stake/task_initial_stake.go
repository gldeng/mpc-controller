package stake

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	addDelegator "github.com/avalido/mpc-controller/tasks/adddelegator"
	"github.com/avalido/mpc-controller/tasks/c2p"
	"github.com/pkg/errors"
)

const (
	taskTypeInitialStake = "initialStake"
)

var (
	_ core.Task = (*InitialStake)(nil)
)

type Status int

type InitialStake struct {
	FlowId   string
	TaskType string
	Quorum   types.QuorumInfo

	C2P          *c2p.C2P
	AddDelegator *addDelegator.AddDelegator

	request *Request

	SubTaskHasError error
	Failed          bool
}

func (t *InitialStake) GetId() string {
	return fmt.Sprintf("%v-%v", t.FlowId, t.TaskType)
}

func (t *InitialStake) FailedPermanently() bool {
	return t.Failed || t.C2P.FailedPermanently() // TODO: Should purely base on sub-tasks? i.e. no own Failed flag
}

func NewInitialStake(request *Request, quorum types.QuorumInfo) (*InitialStake, error) {
	id, err := request.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "failed get StakeRequest hash")
	}
	amount := new(big.Int)
	amount.SetString(request.Amount, 10)
	flowID := "initialStake" + "_" + strconv.FormatUint(request.ReqNo, 10) + "_" + id.String() // TODO: reuse initialStake from common package
	c2pInstance, err := c2p.NewC2P(flowID, quorum, *amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create C2P instance")
	}

	return &InitialStake{
		FlowId:   flowID,
		TaskType: taskTypeInitialStake,
		Quorum:   quorum,
		C2P:      c2pInstance,
		request:  request,
	}, nil
}

func (t *InitialStake) Next(ctx core.TaskContext) ([]core.Task, error) {
	tasks, err := t.run(ctx)
	if err != nil {
		t.SubTaskHasError = err
	}
	return tasks, err
}

func (t *InitialStake) IsDone() bool {
	return t.AddDelegator != nil && t.AddDelegator.IsDone()
}

func (t *InitialStake) IsSequential() bool {
	return t.C2P.IsSequential()
}

func (t *InitialStake) startAddDelegator() error {
	signedImportTx, err := t.C2P.ImportTask.SignedTx()
	nodeID, err := ids.ShortFromPrefixedString(t.request.NodeID, ids.NodeIDPrefix)
	if err != nil {
		return errors.Wrap(err, "failed to convert NodeID")
	}
	stakeParam := addDelegator.StakeParam{
		NodeID:    ids.NodeID(nodeID),
		StartTime: t.request.StartTime,
		EndTime:   t.request.EndTime,
		UTXOs:     signedImportTx.UTXOs(),
	}

	addDelegator, err := addDelegator.NewAddDelegator(t.FlowId, t.Quorum, &stakeParam)
	if err != nil {
		return errors.Wrap(err, "failed to create AddDelegator")
	}
	t.AddDelegator = addDelegator
	return nil
}

func (t *InitialStake) run(ctx core.TaskContext) ([]core.Task, error) {
	if !t.C2P.IsDone() {
		next, err := t.C2P.Next(ctx)
		if t.C2P.IsDone() {
			err = t.startAddDelegator()
			t.logDebug(ctx, "C2P done")
			if err != nil {
				err = errors.Wrap(err, "failed to start AddDelegator")
			}
		}
		if err != nil {
			err = errors.Wrap(err, "failed to run C2P")
		}
		return next, t.failIfErrorf(err, "c2p failed")
	}

	if t.AddDelegator != nil && !t.AddDelegator.IsDone() {
		next, err := t.AddDelegator.Next(ctx)
		if t.AddDelegator.IsDone() {
			t.logDebug(ctx, "AddDelegator task done")
		}
		return next, t.failIfErrorf(err, "AddDelegator failed")
	}
	return nil, t.failIfErrorf(errors.New("invalid state"), "invalid state of composite task")
}

func (t *InitialStake) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *InitialStake) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

// TODO: clear up the below comment on second thought
//func (t *InitialStake) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
//	allFields := make([]logger.Field, 0, len(fields)+3)
//	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
//	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
//	allFields = append(allFields, fields...)
//	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
//	ctx.GetLogger().Error(msg, allFields...)
//}
