package stake

import (
	"encoding/hex"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
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
	SubTaskHasError error
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
		next, err := t.C2P.Next(ctx)
		if len(next) == 1 && next[0] == t.C2P {
			return []core.Task{t}, nil
		}
		return nil, err
	}
	return nil, errors.New("invalid state of composite task")
}
