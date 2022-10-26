package c2p

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/common"
)

type QuorumInfo struct {
	ParticipantPubKeys [][]byte
	PubKey             []byte
}

func (q *QuorumInfo) CChainAddress() common.Address {
	addr, _ := storage.PubKey(q.PubKey).CChainAddress()
	return addr
}

func (q *QuorumInfo) PChainAddress() ids.ShortID {
	addr, _ := storage.PubKey(q.PubKey).PChainAddress()
	return addr
}

type ImportedEvent struct {
	Tx *txs.ImportTx
}

const (
	StatusInit Status = iota
	StatusSignReqSent
	StatusTxSent
	StatusDone
)

type Status int

const (
	SigLength = 65
)
