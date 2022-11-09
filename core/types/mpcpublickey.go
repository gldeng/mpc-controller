package types

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
)

type MpcPublicKey struct {
	GroupId            common.Hash `json:"groupId"`
	GenPubKey          PubKey      `json:"genPubKey"`
	ParticipantPubKeys [][]byte    `json:"participantPubKeys"`
}

func (k *MpcPublicKey) Encode() ([]byte, error) {
	return json.Marshal(k)
}

func (k *MpcPublicKey) Decode(data []byte) error {
	return json.Unmarshal(data, k)
}
