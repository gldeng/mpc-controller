package p2c

import (
	"fmt"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	taskTypeP2C = "p2c"
)

var (
	_ core.Task = (*P2C)(nil)
)

type P2C struct {
	Status   Status
	FlowId   core.FlowId
	TaskType string
	Quorum   types.QuorumInfo

	ToAddress common.Address

	ExportTask      *ExportFromPChain
	ImportTask      *ImportIntoCChain
	SubTaskHasError error
}

// TODO: Support multiple UTXO
func NewP2C(flowId core.FlowId, quorum types.QuorumInfo, utxo avax.UTXO, to common.Address) (*P2C, error) {
	exportTask, err := NewExportFromPChain(flowId, quorum, utxo)
	if err != nil {
		return nil, err
	}
	return &P2C{
		Status:          StatusInit,
		FlowId:          flowId,
		TaskType:        taskTypeP2C,
		Quorum:          quorum,
		ToAddress:       to,
		ExportTask:      exportTask,
		ImportTask:      nil,
		SubTaskHasError: nil,
	}, nil
}

func (t *P2C) GetId() string {
	return fmt.Sprintf("%v-p2c", t.FlowId)
}

func (t *P2C) FailedPermanently() bool {
	return t.ExportTask.FailedPermanently() || (t.ImportTask != nil && t.ImportTask.FailedPermanently())
}

func (t *P2C) IsSequential() bool {
	return !t.ExportTask.IsDone()
}

func (t *P2C) Next(ctx core.TaskContext) ([]core.Task, error) {
	tasks, err := t.run(ctx)
	if err != nil {
		t.SubTaskHasError = err
	}
	return tasks, err
}

func (t *P2C) IsDone() bool {
	return t.ImportTask != nil && t.ImportTask.IsDone()
}

func (t *P2C) startImport() error {
	signedExport, err := t.ExportTask.SignedTx()
	if err != nil {
		return err
	}
	importTask, err := NewImportIntoCChain(t.FlowId, t.Quorum, signedExport, t.ToAddress)
	if err != nil {
		return err
	}
	t.ImportTask = importTask
	return nil
}

func (t *P2C) run(ctx core.TaskContext) ([]core.Task, error) {
	if !t.ExportTask.IsDone() {
		next, err := t.ExportTask.Next(ctx)
		if t.ExportTask.IsDone() {
			t.logDebug(ctx, "export task done")
			err := t.startImport()
			if err != nil {
				t.logError(ctx, "failed to start import task", err)
			}
		}
		return next, err
	}
	if t.ImportTask != nil && !t.ImportTask.IsDone() {
		next, err := t.ImportTask.Next(ctx)
		if err != nil {
			t.SubTaskHasError = err
		}
		if t.ImportTask.IsDone() {
			t.logDebug(ctx, "import task done")
		}
		return next, err
	}
	return nil, errors.New("invalid state of composite task")
}

func (t *P2C) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *P2C) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
