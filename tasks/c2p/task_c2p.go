package c2p

import (
	"github.com/avalido/mpc-controller/pool"
	"github.com/pkg/errors"
	"math/big"
)

var (
	_ pool.Task = (*C2P)(nil)
)

type C2P struct {
	Status Status
	Id     string
	Quorum QuorumInfo

	ExportTask      *ExportFromCChain
	ImportTask      *ImportIntoPChain
	SubTaskHasError error
}

func (t *C2P) RequiresNonce() bool {
	if t.ExportTask.IsDone() {
		return false
	}
	return true
}

func (t *C2P) Next(ctx pool.TaskContext) ([]pool.Task, error) {
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

func (t *C2P) run(ctx pool.TaskContext) ([]pool.Task, error) {
	if !t.ExportTask.IsDone() {
		next, err := t.ExportTask.Next(ctx)
		if len(next) == 1 && next[0] == t.ExportTask {
			return []pool.Task{t}, nil
		}
		if t.ExportTask.IsDone() {
			t.startImport()
		}
		return nil, err
	}
	if t.ImportTask != nil && !t.ImportTask.IsDone() {
		next, err := t.ImportTask.Next(ctx)
		if len(next) == 1 && next[0] == t.ImportTask {
			return []pool.Task{t}, nil
		}
		if err != nil {
			t.SubTaskHasError = err
		}
		return nil, err
	}
	return nil, errors.New("invalid state of composite task")
}

func NewC2P(id string, quorum QuorumInfo, amount big.Int) (*C2P, error) {
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
