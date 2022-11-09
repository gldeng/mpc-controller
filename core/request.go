package core

import (
	"github.com/avalido/mpc-controller/core/types"
)

type Request interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
	Hash() (types.RequestHash, error)
}
