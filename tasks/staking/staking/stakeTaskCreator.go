package staking

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/pkg/errors"
	"math/big"
)

type StakeTaskCreator struct {
	TaskID string
	*storage.StakeRequest
	chain.NetworkContext
	Nonce uint64
}

func (s *StakeTaskCreator) CreateStakeTask() (*StakeTask, error) {
	nodeID, err := ids.ShortFromPrefixedString(s.NodeID, ids.NodeIDPrefix)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reqNo := s.ReqNo

	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(s.Amount, 10)

	startTime := big.NewInt(s.StartTime)
	endTIme := big.NewInt(s.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New("invalid uint64")
	}
	task, err := NewStakeTask(s.TaskID, s.NetworkContext, reqNo, s.GenPubKey, s.Nonce, ids.NodeID(nodeID), nAVAXAmount.Uint64(), startTime.Uint64(), endTIme.Uint64(), cchain.BaseFeeGwei)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return task, nil
}
