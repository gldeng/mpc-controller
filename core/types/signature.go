package types

import "github.com/avalido/mpc-controller/utils/bytes"

type Signature [65]byte

func (s *Signature) FromHex(hex string) *Signature {
	*s = bytes.BytesTo65Bytes(bytes.HexToBytes(hex))
	return s
}

func (s *Signature) String() string {
	return bytes.Bytes65ToHex(*s)
}
