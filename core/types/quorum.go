package types

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
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

func (q *QuorumInfo) CompressKeys() ([]string, string, error) {
	partiPubKeys, err := q.ParticipantPubKeys.CompressPubKeyHexs()
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to compress participant public keys")
	}

	genPubKey, err := q.PubKey.CompressPubKeyHex()
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to compress generated public key")
	}

	return partiPubKeys, genPubKey, nil
}
