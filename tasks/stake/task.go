package stake

import (
	"encoding/hex"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	addDelegator "github.com/avalido/mpc-controller/tasks/adddelegator"
	"github.com/avalido/mpc-controller/tasks/c2p"
	"github.com/pkg/errors"
	"math/big"
)

var (
	_ core.Task = (*InitialStake)(nil)
)

type Status int

type InitialStake struct {
	Id     string
	Quorum types.QuorumInfo

	C2P          *c2p.C2P
	AddDelegator *addDelegator.AddDelegator

	request *Request

	SubTaskHasError error
	Failed          bool
}

func (t *InitialStake) GetId() string {
	return fmt.Sprintf("Stake(%v)", t.Id)
}

func (t *InitialStake) FailedPermanently() bool {
	return t.Failed || t.C2P.FailedPermanently()
}

func NewInitialStake(request *Request, quorum types.QuorumInfo) (*InitialStake, error) {
	id, err := request.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "failed get StakeRequest hash")
	}
	amount := new(big.Int)
	amount.SetString(request.Amount, 10)
	c2pInstance, err := c2p.NewC2P(hex.EncodeToString(id[:]), quorum, *amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create C2P instance")
	}

	return &InitialStake{
		Id:      hex.EncodeToString(id[:]),
		Quorum:  quorum,
		C2P:     c2pInstance,
		request: request,
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

func (t *InitialStake) RequiresNonce() bool {
	return t.C2P.RequiresNonce()
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

	addDelegator, err := addDelegator.NewAddDelegator(t.Id, t.Quorum, &stakeParam)
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
			ctx.GetLogger().Debug(fmt.Sprintf("%v C2P done", t.Id))
			if err != nil {
				ctx.GetLogger().Errorf("Failed to start AddDelegator, error:%+v", err)
			}
		}
		if err != nil {
			ctx.GetLogger().Errorf("Failed to run C2P, error: %v", err)
		}
		return next, err
	}

	if t.AddDelegator != nil && !t.AddDelegator.IsDone() {
		next, err := t.AddDelegator.Next(ctx)
		if t.AddDelegator.IsDone() {
			ctx.GetLogger().Debug(fmt.Sprintf("%v added delegator", t.Id))
			if err != nil {
				ctx.GetLogger().Errorf("%v AddDelegator got error:%+v", t.Id, err)
			}
		}
		return next, err
	}
	return nil, t.failIfError(errors.New("invalid state"), "invalid state of composite task")
}

func (t *InitialStake) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, msg)
}
