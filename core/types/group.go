package types

import (
	"encoding/json"
	"math/big"
)

type Group struct {
	GroupId [32]byte
	Index   *big.Int
}

func (g *Group) Encode() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Group) Decode(data []byte) error {
	return json.Unmarshal(data, g)
}

type Groups []Group

func (g Groups) Encode() ([]byte, error) {
	return json.Marshal(g)
}

func (g Groups) Decode(data []byte) error {
	return json.Unmarshal(data, g)
}
