package c2p

import (
	"fmt"
	"math/big"

	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/pkg/errors"
)

const (
	taskTypeC2P = "c2p"
)

var (
	_ core.Task = (*C2P)(nil)
)

type C2P struct {
	Status   Status
	FlowId   core.FlowId
	TaskType string
	Quorum   types.QuorumInfo

	ExportTask      *ExportFromCChain
	ImportTask      *ImportIntoPChain
	SubTaskHasError error
}

func NewC2P(flowId core.FlowId, quorum types.QuorumInfo, amount big.Int) (*C2P, error) {
	exportTask, err := NewExportFromCChain(flowId, quorum, amount)
	if err != nil {
		return nil, err
	}
	return &C2P{
		Status:     StatusInit,
		FlowId:     flowId,
		TaskType:   taskTypeC2P,
		Quorum:     quorum,
		ExportTask: exportTask,
	}, nil
}

func (t *C2P) GetId() string {
	return fmt.Sprintf("%v-c2p", t.FlowId)
}

func (t *C2P) FailedPermanently() bool {
	return t.ExportTask.FailedPermanently() || (t.ImportTask != nil && t.ImportTask.FailedPermanently())
}

func (t *C2P) IsSequential() bool {
	return !t.ExportTask.IsDone()
}

func (t *C2P) Next(ctx core.TaskContext) ([]core.Task, error) {
	tasks, err := t.run(ctx)
	if err != nil {
		t.SubTaskHasError = err
	}
	return tasks, err
}

func (t *C2P) IsDone() bool {
	return t.ImportTask != nil && t.ImportTask.IsDone()
}

func (t *C2P) startImport() error {
	signedExport, err := t.ExportTask.SignedTx()
	if err != nil {
		return err
	}
	importTask, err := NewImportIntoPChain(t.FlowId, t.Quorum, signedExport)
	if err != nil {
		return err
	}
	t.ImportTask = importTask
	return nil
}

func (t *C2P) run(ctx core.TaskContext) ([]core.Task, error) {
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

func (t *C2P) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *C2P) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
