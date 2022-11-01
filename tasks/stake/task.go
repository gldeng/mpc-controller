package stake

import (
	"encoding/hex"
	"fmt"
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

const (
	StatusInit Status = iota
	StatusSignReqSent
	StatusTxSent
	StatusDone
)

type InitialStake struct {
	Status Status
	Id     string
	Quorum types.QuorumInfo

	C2P             *c2p.C2P
	AddDelegator    *addDelegator.AddDelegator
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
		Status: StatusInit,
		Id:     hex.EncodeToString(id[:]),
		Quorum: quorum,
		C2P:    c2pInstance,
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
	return t.C2P.IsDone()
}

func (t *InitialStake) RequiresNonce() bool {
	return t.C2P.RequiresNonce()
}

func (t *InitialStake) run(ctx core.TaskContext) ([]core.Task, error) {
	// TODO: Add AddDelegator Tx
	if !t.C2P.IsDone() {
		return t.C2P.Next(ctx)
	}
	if t.C2P.IsDone() {
		signedImportTx, err := t.C2P.ImportTask.SignedTx()
		if err != nil {
			return nil, t.failIfError(err, "failed to get signed ImportTx")
		}
		addDelegator, err := addDelegator.NewAddDelegator(nil, t.Id, t.Quorum, signedImportTx) // todo: give request params
		if err != nil {
			return nil, t.failIfError(err, "failed to create AddDelegator task")
		}
		t.AddDelegator = addDelegator
		return t.AddDelegator.Next(ctx)
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
