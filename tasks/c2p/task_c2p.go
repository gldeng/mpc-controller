package c2p

import (
	"fmt"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/pkg/errors"
	"math/big"
)

var (
	_ core.Task = (*C2P)(nil)
)

type C2P struct {
	Status Status
	Id     string
	Quorum types.QuorumInfo

	ExportTask      *ExportFromCChain
	ImportTask      *ImportIntoPChain
	SubTaskHasError error
}

func (t *C2P) GetId() string {
	return fmt.Sprintf("C2P(%v)", t.Id)
}

func (t *C2P) FailedPermanently() bool {
	return t.ExportTask.FailedPermanently() || (t.ImportTask != nil && t.ImportTask.FailedPermanently())
}

func (t *C2P) IsSequential() bool {
	if t.ExportTask.IsDone() {
		return false
	}
	return true
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
	importTask, err := NewImportIntoPChain(t.Id, t.Quorum, signedExport)
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
			ctx.GetLogger().Debugf("%v imported", t.Id)
		}
		return next, err
	}
	return nil, errors.New("invalid state of composite task")
}

func NewC2P(id string, quorum types.QuorumInfo, amount big.Int) (*C2P, error) {
	exportTask, err := NewExportFromCChain(id, quorum, amount)
	if err != nil {
		return nil, err
	}
	return &C2P{
		Status:     StatusInit,
		Id:         id,
		Quorum:     quorum,
		ExportTask: exportTask,
	}, nil
}
