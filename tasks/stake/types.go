package stake

import (
	"encoding/json"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ core.Request = (*Request)(nil)
)

type Request struct {
	// The request number of the request
	ReqNo uint64 `json:"reqNo"`
	// The tx hash that emits the event
	TxHash common.Hash `json:"txHash"`
	// The public key of the MPC wallet to be used
	PubKey []byte `json:"pubKey"`
	// The NodeID that we need to stake to
	NodeID string `json:"nodeID"`
	// The amount to be staked
	Amount string `json:"amount"`
	// The StartTime of the stake period
	StartTime uint64 `json:"startTime"`
	// The EndTime of the stake period
	EndTime uint64 `json:"endTime"`
}

type P2CRequest struct {
	// The TxID of the transaction that contains the UTXO on P-Chain
	TxID ids.ID `json:"txID"`
	// The index of the UTXO in the transaction
	OutputIndex uint32 `json:"outputIndex"`
	// The Address that will receive the AVAX on the C-Chain
	DestinationAddress common.Address `json:"outputIndex"`
}

func (r *Request) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Request) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r Request) Hash() (types.RequestHash, error) {
	data, err := r.Encode()
	if err != nil {
		return [32]byte{}, err
	}
	hash := types.RequestHash(hash256.FromBytes(data))
	hash.SetTaskType(types.TaskTypStake)
	return hash, nil
}
