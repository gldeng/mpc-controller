package recovery

import (
	"encoding/json"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
)

const (
	StatusInit Status = iota
	StatusSignReqSent
	StatusSignedTxReady
	StatusTxSent
	StatusNotNeeded
	StatusDone
)

type Status int

type Request struct {
	// The RequestHash of the failed request
	OriginalRequestHash types.RequestHash `json:"originalRequestHash"`
	ExportTxID          ids.ID            `json:"exportTxID"`
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
	hash.SetTaskType(types.TaskTypRecover)
	return hash, nil
}
