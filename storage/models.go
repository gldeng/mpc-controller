package storage

import (
	"bytes"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/common"
)

// Key prefixes
var (
	KeyPrefixGroup              KeyPrefix = []byte("group-")
	KeyPrefixParticipant        KeyPrefix = []byte("parti-")
	KeyPrefixGeneratedPublicKey KeyPrefix = []byte("genPubKey-")
)

const (
	PubKeyLength = 64
)

type PubKey [PubKeyLength]byte

// Models

type Group struct {
	ID    common.Hash `json:"id"`
	Group []PubKey    `json:"group"`
}

type Participant struct {
	PubKey  common.Hash `json:"pubKey"`
	GroupId common.Hash `json:"groupId"`
	Index   uint64      `json:"index"`
}

type GeneratedPublicKey struct {
	GenPubKey PubKey      `json:"genPubKey"`
	GroupId   common.Hash `json:"groupId"`
}

// Model keys

// Key format: KeyPrefixGroup+ID
func (m *Group) Key() []byte {
	keyPayload := m.ID
	return Key(KeyPrefixGroup, KeyPayload(keyPayload))
}

// Key format: KeyPrefixParticipant+Hash(PubKey+GroupId)
func (m *Participant) Key() []byte {
	keyPayload := hash256.FromBytes(bytes.Join([][]byte{m.PubKey.Bytes(), m.GroupId.Bytes()}, []byte("")))
	return Key(KeyPrefixParticipant, KeyPayload(keyPayload))
}

// Key format: KeyPrefixGeneratedPublicKey+Hash(GenPubKey)
func (m *GeneratedPublicKey) Key() []byte {
	keyPayload := hash256.FromBytes(m.GenPubKey[:])
	return Key(KeyPrefixGeneratedPublicKey, KeyPayload(keyPayload))
}
