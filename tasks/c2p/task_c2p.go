package c2p

import (
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

func (t *C2P) FailedPermanently() bool {
	return t.ExportTask.FailedPermanently() || (t.ImportTask != nil && t.ImportTask.FailedPermanently())
}

func (t *C2P) RequiresNonce() bool {
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
		if len(next) == 1 && next[0] == t.ExportTask {
			return []core.Task{t}, nil
		}
		if t.ExportTask.IsDone() {
			err := t.startImport()
			ctx.GetLogger().ErrorOnError(err, "failed to start import")
		}
		return nil, err
	}
	if t.ImportTask != nil && !t.ImportTask.IsDone() {
		next, err := t.ImportTask.Next(ctx)
		if len(next) == 1 && next[0] == t.ImportTask {
			return []core.Task{t}, nil
		}
		if err != nil {
			t.SubTaskHasError = err
		}
		if t.ImportTask.IsDone() {
			ctx.GetLogger().Debug("imported")
		}
		return nil, err
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
