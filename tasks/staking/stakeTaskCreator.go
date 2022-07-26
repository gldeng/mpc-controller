package staking

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/pkg/errors"
	"math/big"
)

type StakeTaskCreator struct {
	TaskID string
	*contract.MpcManagerStakeRequestStarted
	chain.NetworkContext
	PubKeyHex string
	Nonce     uint64
}

func (s *StakeTaskCreator) CreateStakeTask() (*StakeTask, error) {
	nodeID, err := ids.ShortFromPrefixedString(s.NodeID, ids.NodeIDPrefix)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pubKey, err := crypto.UnmarshalPubKeyHex(s.PubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	requestID := s.RequestId.Uint64()

	nAVAXAmount := new(big.Int).Div(s.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !s.StartTime.IsUint64() || !s.EndTime.IsUint64() {
		return nil, errors.New("invalid uint64")
	}
	task, err := NewStakeTask(s.TaskID, s.NetworkContext, requestID, *pubKey, s.Nonce, ids.NodeID(nodeID), nAVAXAmount.Uint64(), s.StartTime.Uint64(), s.EndTime.Uint64(), cchain.BaseFeeGwei)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return task, nil
}
