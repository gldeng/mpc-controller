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

type stakeTaskCreator struct {
	*contract.MpcManagerStakeRequestStarted
	chain.NetworkContext
	PubKeyHex string
	Nonce     uint64
}

func (m *stakeTaskCreator) createStakeTask() (*StakeTask, error) {
	nodeID, err := ids.ShortFromPrefixedString(m.NodeID, constants.NodeIDPrefix)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pubKey, err := crypto.UnmarshalPubKeyHex(m.PubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	baseFeeGwei := uint64(300) // TODO: It should be given by the contract

	nAVAXAmount := new(big.Int).Div(m.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !m.StartTime.IsUint64() || !m.EndTime.IsUint64() {
		return nil, errors.New("invalid uint64")
	}
	task, err := NewStakeTask(m.NetworkContext, *pubKey, m.Nonce, nodeID, nAVAXAmount.Uint64(), m.StartTime.Uint64(), m.EndTime.Uint64(), baseFeeGwei)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return task, nil
}
