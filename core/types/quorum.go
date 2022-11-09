package types

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
)

type QuorumInfo struct {
	ParticipantPubKeys [][]byte
	PubKey             []byte
}

func (q *QuorumInfo) CChainAddress() common.Address {
	addr, _ := PubKey(q.PubKey).CChainAddress()
	return addr
}

func (q *QuorumInfo) PChainAddress() ids.ShortID {
	addr, _ := PubKey(q.PubKey).PChainAddress()
	return addr
}
