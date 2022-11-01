package core

import "github.com/avalido/mpc-controller/storage"

type Request interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
	Hash() (storage.RequestHash, error)
}
