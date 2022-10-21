package c2p

import "github.com/ava-labs/avalanchego/vms/platformvm/txs"

type QuorumInfo struct {
	ParticipantPubKeys [][]byte
	PubKey             []byte
}

type ImportedEvent struct {
	Tx *txs.ImportTx
}
