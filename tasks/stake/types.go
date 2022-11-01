package stake

import (
	"encoding/json"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ core.Request = (*Request)(nil)
)

type Request struct {
	ReqNo     uint64      `json:"reqNo"`
	TxHash    common.Hash `json:"txHash"`
	PubKey    []byte      `json:"pubKey"`
	NodeID    string      `json:"nodeID"`
	Amount    string      `json:"amount"`
	StartTime uint64      `json:"startTime"`
	EndTime   uint64      `json:"endTime"`
}

func (r *Request) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Request) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r Request) Hash() (storage.RequestHash, error) {
	data, err := r.Encode()
	if err != nil {
		return [32]byte{}, err
	}
	hash := storage.RequestHash(hash256.FromBytes(data))
	hash.SetTaskType(storage.TaskTypStake)
	return hash, nil
}
