package staking

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/pkg/errors"
	"math/big"
)

type StakeTaskCreator struct {
	*contract.MpcManagerStakeRequestStarted
	chain.NetworkContext
	PubKeyHex string
	Nonce     uint64
}

func (s *StakeTaskCreator) CreateStakeTask() (*StakeTask, error) {
	nodeID, err := ids.ShortFromPrefixedString(s.NodeID, constants.NodeIDPrefix)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pubKey, err := crypto.UnmarshalPubKeyHex(s.PubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	baseFeeGwei := uint64(300) // TODO: It should be given by the contract

	nAVAXAmount := new(big.Int).Div(s.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !s.StartTime.IsUint64() || !s.EndTime.IsUint64() {
		return nil, errors.New("invalid uint64")
	}
	task, err := NewStakeTask(s.NetworkContext, *pubKey, s.Nonce, nodeID, nAVAXAmount.Uint64(), s.StartTime.Uint64(), s.EndTime.Uint64(), baseFeeGwei)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return task, nil
}
