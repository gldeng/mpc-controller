package types

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
)

type QuorumInfo struct {
	ParticipantPubKeys PubKeys
	PubKey             PubKey
}

func (q *QuorumInfo) CChainAddress() common.Address {
	addr, _ := q.PubKey.CChainAddress()
	return addr
}

func (q *QuorumInfo) PChainAddress() ids.ShortID {
	addr, _ := q.PubKey.PChainAddress()
	return addr
}
