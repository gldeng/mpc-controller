package p2c

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	taskTypeMoveBucket = "moveBucket"
)

var (
	_ core.Task = (*MoveBucket)(nil)
)

type MoveBucket struct {
	FlowId   core.FlowId
	TaskType string
	Quorum   types.QuorumInfo

	request *MoveBucketRequest

	P2C *P2C

	SubTaskHasError error
	Failed          bool
}

func (t *MoveBucket) GetId() string {
	return fmt.Sprintf("%v-%v", t.FlowId, t.TaskType)
}

func (t *MoveBucket) FailedPermanently() bool {
	return t.Failed || t.P2C.FailedPermanently() // TODO: Should purely base on sub-tasks? i.e. no own Failed flag
}

func (t *MoveBucket) IsDone() bool {
	return t.P2C != nil && t.P2C.IsDone()
}

func (t *MoveBucket) IsSequential() bool {
	return t.P2C.IsSequential()
}

func NewMoveBucket(request *MoveBucketRequest, quorum types.QuorumInfo) (*MoveBucket, error) {
	id, err := request.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "failed get MoveBucketRequest hash")
	}
	flowID := core.FlowId{
		Tag:         fmt.Sprintf("move_bucket_%v_%v_%v", request.Bucket.UtxoType.String(), request.Bucket.StartTimestamp, request.Bucket.EndTimestamp),
		RequestHash: id,
	}
	return &MoveBucket{
		FlowId:          flowID,
		TaskType:        taskTypeMoveBucket,
		Quorum:          quorum,
		request:         request,
		SubTaskHasError: nil,
		Failed:          false,
	}, nil
}

func (t *MoveBucket) Next(ctx core.TaskContext) ([]core.Task, error) {
	tasks, err := t.run(ctx)
	if err != nil {
		t.SubTaskHasError = err
	}
	return tasks, err
}

func (t *MoveBucket) run(ctx core.TaskContext) ([]core.Task, error) {
	if t.P2C == nil {
		toAddress := common.Address{}
		if t.request.Bucket.UtxoType == types.Principal {
			addr, err := ctx.PrincipalTreasuryAddress(nil)
			if err != nil {
				return nil, t.failIfErrorf(err, ErrMsgFailedToGetPrincipalTreasuryAddress)
			}
			toAddress = addr
		}
		if t.request.Bucket.UtxoType == types.Reward {
			addr, err := ctx.RewardTreasuryAddress(nil)
			if err != nil {
				return nil, t.failIfErrorf(err, ErrMsgFailedToGetRewardTreasuryAddress)
			}
			toAddress = addr
		}
		p2cInstance, err := NewP2C(t.FlowId, t.Quorum, t.request.Bucket.Utxos, toAddress)
		if err != nil {
			return nil, t.failIfErrorf(err, ErrMsgFailedToCreateP2CTask)
		}
		t.P2C = p2cInstance
	}

	if !t.P2C.IsDone() {
		next, err := t.P2C.Next(ctx)
		if err != nil {
			t.logError(ctx, "Failed to run P2C", err)
		}
		return next, t.failIfErrorf(err, "p2c failed")
	}

	return nil, t.failIfErrorf(errors.New("invalid state"), "invalid state of composite task")
}

func (t *MoveBucket) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *MoveBucket) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *MoveBucket) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
