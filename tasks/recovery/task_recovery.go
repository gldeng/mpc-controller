package recovery

import (
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/c2p"
	"github.com/avalido/mpc-controller/tasks/p2c"
	myAvax "github.com/avalido/mpc-controller/utils/txs/avax"
	"github.com/pkg/errors"
)

const (
	taskTypeRecovery = "recovery"
)

var (
	_ core.Task = (*Recovery)(nil)
)

type Recovery struct {
	FlowId   core.FlowId
	Status   Status
	TaskType string

	request *Request

	Utxo            *avax.UTXO
	Import          *c2p.ImportIntoPChain // Complete atomic tx if not already
	P2C             *p2c.P2C
	SubTaskHasError error

	Quorum types.QuorumInfo
	Failed bool
}

func (t *Recovery) GetId() string {
	return fmt.Sprintf("%v-recovery", t.FlowId)
}

func (t *Recovery) FailedPermanently() bool {
	return t.Failed || (
		t.Import != nil && t.Import.FailedPermanently()) || (
		t.P2C != nil && t.P2C.FailedPermanently())
}

func (t *Recovery) IsSequential() bool {
	return true
}

func (t *Recovery) IsDone() bool {
	return t.Status == StatusDone
}

func NewRecovery(request *Request, quorum types.QuorumInfo) (*Recovery, error) {
	id, err := request.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "failed get StakeRequest hash")
	}

	flowID := core.FlowId{
		Tag:         "recovery" + "_" + request.OriginalRequestHash.String(),
		RequestHash: id,
	}

	return &Recovery{
		FlowId:   flowID,
		TaskType: taskTypeRecovery,
		Quorum:   quorum,
		request:  request,
	}, nil
}

func (t *Recovery) Next(ctx core.TaskContext) ([]core.Task, error) {
	switch t.Status {
	case StatusInit:
		err := t.initialize(ctx)
		if err != nil {
			return nil, err
		}
	case StatusRunningImport:
		if !t.Import.IsDone() {
			next, err := t.Import.Next(ctx)
			if err != nil {
				t.SubTaskHasError = err
			}
			if t.Import.IsDone() {
				t.logDebug(ctx, "import task done")
				utxos := myAvax.UTXOsFromTransferableOutputs(*t.Import.TxID, t.Import.Tx.Outputs())
				err = t.createP2C(ctx, *utxos[0])
				if err != nil {
					return next, err
				}
			}
			return next, err
		}
		return nil, nil
	case StatusRunningP2C:
		if !t.P2C.IsDone() {
			next, err := t.P2C.Next(ctx)
			if err != nil {
				t.SubTaskHasError = err
			}
			if t.P2C.IsDone() {
				t.logDebug(ctx, "p2c task done")
				t.Status = StatusDone
			}
			return next, err
		}
		return nil, nil
	}
	return nil, nil
}

func (t *Recovery) initialize(ctx core.TaskContext) error {
	txId, _ := ctx.GetTxIndex().GetTxByType(t.request.OriginalRequestHash, core.TxTypeAddDelegator)
	if txId != ids.Empty {
		t.Status = StatusDone
		return nil
	}
	txId, _ = ctx.GetTxIndex().GetTxByType(t.request.OriginalRequestHash, core.TxTypeImportP)
	if txId != ids.Empty {
		tx, err := ctx.GetPChainTx(txId)
		if err != nil {
			return t.failIfErrorf(err, ErrMsgFailedToRetrieveTx)
		}
		t.createP2C(ctx, *tx.UTXOs()[0])
	}
	if t.request.ExportTxID != ids.Empty {
		originalExportTx, err := ctx.GetCChainTx(t.request.ExportTxID)
		if err != nil {
			return t.failIfErrorf(err, ErrMsgFailedToRetrieveTx)
		}
		return t.createImport(ctx, originalExportTx)
	}
	err := errors.New("invalid task")
	return t.failIfErrorf(err, ErrMsgInvalidTaskk)
}

func (t *Recovery) createImport(ctx core.TaskContext, originalExportTx *evm.Tx) error {
	importTask, err := c2p.NewImportIntoPChain(t.FlowId, t.Quorum, originalExportTx)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCreateSubTask)
	}
	t.Import = importTask
	t.Status = StatusRunningImport
	return nil
}

func (t *Recovery) createP2C(ctx core.TaskContext, utxo avax.UTXO) error {
	toAddress, err := ctx.PrincipalTreasuryAddress(nil)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToGetPrincipalTreasuryAddress)
	}
	p2cTask, err := p2c.NewP2C(t.FlowId, t.Quorum, utxo, toAddress)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCreateSubTask)
	}
	t.P2C = p2cTask
	t.Status = StatusRunningP2C
	return nil
}

func (t *Recovery) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *Recovery) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *Recovery) logInfo(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Info(msg, allFields...)
}

func (t *Recovery) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
