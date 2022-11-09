package types

import (
	"encoding/json"
	"math/big"
)

type Group struct {
	GroupId          [32]byte
	Index            *big.Int
	MemberPublicKeys [][]byte
}

func (g *Group) Encode() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Group) Decode(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *Group) ParticipantID() ParticipantId {
	groupIdBig := new(big.Int).SetBytes(g.GroupId[:])
	indexBig := new(big.Int).SetUint64(g.Index.Uint64())
	partiIdBig := new(big.Int).Or(groupIdBig, indexBig)
	var partiId [32]byte
	copy(partiId[:], partiIdBig.Bytes())
	return partiId
}

type Groups []Group

func (g Groups) Encode() ([]byte, error) {
	return json.Marshal(g)
}

func (g Groups) Decode(data []byte) error {
	return json.Unmarshal(data, g)
}
