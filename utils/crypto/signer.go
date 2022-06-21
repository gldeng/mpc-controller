package crypto

import (
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
)

var _ Signer_ = (*SECP256K1RSigner)(nil)

type Signer_ interface {
	Sign(message []byte) ([]byte, error)
	SignHash(hash []byte) ([]byte, error)

	Verify(message, signature []byte) bool
	VerifyHash(hash, signature []byte) bool

	PrivateKey() avaCrypto.PrivateKey
	PublicKey() avaCrypto.PublicKey
	Address() ids.ShortID
}

type SECP256K1RSigner struct {
	privKey avaCrypto.PrivateKey
	pubKey  avaCrypto.PublicKey
	addr    ids.ShortID
}

func NewSECP256K1RSigner() (Signer_, error) {
	factory := &avaCrypto.FactorySECP256K1R{}
	s := SECP256K1RSigner{}
	privKey, err := factory.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	s.privKey = privKey
	s.pubKey = s.privKey.PublicKey()
	s.addr = s.pubKey.Address()
	return &s, nil
}

func ToSECP256K1RSigner(b []byte) (Signer_, error) {
	factory := &avaCrypto.FactorySECP256K1R{}
	privKey, err := factory.ToPrivateKey(b)
	if err != nil {
		return nil, err
	}

	s := SECP256K1RSigner{}
	s.privKey = privKey
	s.pubKey = s.privKey.PublicKey()
	s.addr = s.pubKey.Address()
	return &s, nil
}

func (s *SECP256K1RSigner) Sign(message []byte) ([]byte, error) {
	return s.privKey.Sign(message)
}

func (s *SECP256K1RSigner) SignHash(hash []byte) ([]byte, error) {
	return s.privKey.SignHash(hash)
}

func (s *SECP256K1RSigner) Verify(message, signature []byte) bool {
	return s.pubKey.Verify(message, signature)
}

func (s *SECP256K1RSigner) VerifyHash(hash, signature []byte) bool {
	return s.pubKey.VerifyHash(hash, signature)
}

func (s *SECP256K1RSigner) PrivateKey() avaCrypto.PrivateKey {
	return s.privKey
}

func (s *SECP256K1RSigner) PublicKey() avaCrypto.PublicKey {
	return s.pubKey
}

func (s *SECP256K1RSigner) Address() ids.ShortID {
	return s.addr
}
