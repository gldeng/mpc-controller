package p2c

import (
	"encoding/json"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
)

type ImportedEvent struct {
	Tx *txs.ImportTx
}

const (
	StatusInit Status = iota
	StatusSignReqSent
	StatusSignedTxReady
	StatusTxSent
	StatusDone
)

type Status int

const (
	SigLength = 65
)

var (
	_ core.Request = (*MoveBucketRequest)(nil)
)

type MoveBucketRequest struct {
	// The UTXO Bucket to move
	Bucket types.UtxoBucket `json:"bucket"`
}

func (m *MoveBucketRequest) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MoveBucketRequest) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *MoveBucketRequest) Hash() (types.RequestHash, error) {
	data, err := m.Encode()
	if err != nil {
		return [32]byte{}, err
	}
	hash := types.RequestHash(hash256.FromBytes(data))
	hash.SetTaskType(types.TaskTypP2C)
	return hash, nil
}
