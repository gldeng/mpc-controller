package failure

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"time"
)

const (
	taskTypeFailureReport = "FailureReport"
	checkTxDelay          = 30 * time.Second
)

var (
	_ core.Task = (*Report)(nil)
)

type Report struct {
	FlowId   core.FlowId
	TaskType string
	Status   Status

	ParticipantId     types.ParticipantId
	FailedRequestHash types.RequestHash
	ExtraData         []byte
	Failed            bool
	StartTime         time.Time
	TxSentTime        *time.Time

	TxHash common.Hash
}

func (t *Report) GetId() string {
	return fmt.Sprintf("FailureReport(%x)", t.FlowId.RequestHash)
}

func (t *Report) FailedPermanently() bool {
	return t.Failed
}

func NewFailureReport(flowId core.FlowId, participantId types.ParticipantId, requestHash types.RequestHash, data []byte) *Report {
	return &Report{
		FlowId:            flowId,
		TaskType:          taskTypeFailureReport,
		Status:            StatusInit,
		ParticipantId:     participantId,
		FailedRequestHash: requestHash,
		ExtraData:         data,
		Failed:            false,
		StartTime:         time.Now(),
	}
}

func (t *Report) Next(ctx core.TaskContext) ([]core.Task, error) {
	if time.Now().Sub(t.StartTime) < core.DefaultParameters.ReportFailureDelay {
		return nil, nil
	}
	switch t.Status {
	case StatusInit:

		txHash, err := ctx.ReportRequestFailed(ctx.GetMyTransactSigner(), t.ParticipantId, t.FailedRequestHash, t.ExtraData)

		if err != nil {
			return nil, t.failIfErrorf(err, ErrMsgReportFailed)
		}
		t.TxHash = *txHash
		t.Status = StatusTxSent
		sentTime := time.Now()
		t.TxSentTime = &sentTime
		t.logDebug(ctx, "sent failure report tx")
	case StatusTxSent:
		_, err := ctx.CheckEthTx(t.TxHash)
		if err != nil {
			if time.Now().Sub(*t.TxSentTime) > checkTxDelay {
				t.logError(ctx, ErrMsgCheckTxStatus, err)
				return nil, t.failIfErrorf(err, ErrMsgCheckTxStatus)
			} else {
				// If within the grace period, the tx may have not been received.
				return nil, nil
			}
		}

		t.Status = StatusDone
		t.logDebug(ctx, "failure reported")
	}
	return nil, nil
}

func (t *Report) IsDone() bool {
	return t.Status == StatusDone
}

func (t *Report) IsSequential() bool {
	return true
}

func (t *Report) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *Report) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *Report) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
