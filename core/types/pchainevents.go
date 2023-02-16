package types

import (
	"encoding/json"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

type UtxoType int

const (
	Unknown UtxoType = iota
	Principal
	Reward
)

func (t UtxoType) String() string {
	switch t {
	case Principal:
		return "Principal"
	case Reward:
		return "Reward"
	default:
		return "Unknown"
	}
}

type UtxoBucket struct {
	StartTimestamp uint64
	EndTimestamp   uint64
	PublicKey      []byte
	Utxos          []*avax.UTXO
	UtxoType       UtxoType
}

type utxoBucketJSON struct {
	StartTimestamp uint64   `json:"startTimestamp"`
	EndTimestamp   uint64   `json:"endTimestamp"`
	PublicKey      []byte   `json:"publicKey"`
	Utxos          []string `json:"utxos"`
	UtxoType       UtxoType `json:"utxoType"`
}

func (b *UtxoBucket) MarshalJSON() ([]byte, error) {
	var utxoHexes []string
	for _, utxo := range b.Utxos {
		bytes, err := platformvm.Codec.Marshal(txs.CodecVersion, utxo)
		if err != nil {
			return nil, err
		}
		utxoHex, err := formatting.Encode(formatting.Hex, bytes)
		utxoHexes = append(utxoHexes, utxoHex)
	}
	return json.Marshal(&utxoBucketJSON{
		StartTimestamp: b.StartTimestamp,
		EndTimestamp:   b.EndTimestamp,
		PublicKey:      b.PublicKey,
		Utxos:          utxoHexes,
		UtxoType:       b.UtxoType,
	})
}

func (b *UtxoBucket) UnmarshalJSON(data []byte) error {
	sub := &utxoBucketJSON{}
	if err := json.Unmarshal(data, &sub); err != nil {
		return err
	}
	var utxos []*avax.UTXO
	for _, utxoHex := range sub.Utxos {
		bytes, err := formatting.Decode(formatting.Hex, utxoHex)
		if err != nil {
			return err
		}
		utxo := &avax.UTXO{}
		_, _ = platformvm.Codec.Unmarshal(bytes, utxo)
		utxos = append(utxos, utxo)
	}

	b.StartTimestamp = sub.StartTimestamp
	b.EndTimestamp = sub.EndTimestamp
	b.PublicKey = sub.PublicKey
	b.UtxoType = sub.UtxoType
	b.Utxos = utxos

	return nil
}
