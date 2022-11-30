package c2p

import (
	"fmt"
	"math/big"

	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*C2P)(nil)
)

type C2P struct {
	Status Status
	FlowId string
	Quorum types.QuorumInfo

	ExportTask      *ExportFromCChain
	ImportTask      *ImportIntoPChain
	SubTaskHasError error
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
			err := t.startImport()
			if err != nil {
				ctx.GetLogger().Errorf("Failed to start import, error:%+v", err)
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
			ctx.GetLogger().Debugf("%v imported", t.FlowId)
		}
		return next, err
	}
	return nil, errors.New("invalid state of composite task")
}

func NewC2P(FlowId string, quorum types.QuorumInfo, amount big.Int) (*C2P, error) {
	exportTask, err := NewExportFromCChain(FlowId, quorum, amount)
	if err != nil {
		return nil, err
	}
	return &C2P{
		Status:     StatusInit,
		FlowId:     FlowId,
		Quorum:     quorum,
		ExportTask: exportTask,
	}, nil
}
