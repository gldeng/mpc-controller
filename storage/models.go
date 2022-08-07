package storage

import (
	"encoding/binary"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/common"
)

// Key prefixes
var (
	KeyPrefixGroup              KeyPrefix = []byte("group")
	KeyPrefixParticipant        KeyPrefix = []byte("parti")
	KeyPrefixGeneratedPublicKey KeyPrefix = []byte("genPubKey")

	KeyPrefixStakeRequest KeyPrefix = []byte("stakeRequest")
)

type PubKey []byte

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

type StakeRequest struct {
	ReqNo     string      `json:"reqNo"`
	TxHash    common.Hash `json:"txHash"`
	GenPubKey common.Hash `json:"genPubKey"`
	NodeID    string      `json:"nodeID"`
	Amount    string      `json:"amount"`
	StartTime string      `json:"startTime"`
	EndTime   string      `json:"endTime"`
}

// Model keys

// Key format: KeyPrefixGroup+"-"+ID
func (m *Group) Key() []byte {
	keyPayload := m.ID
	return Key(KeyPrefixGroup, KeyPayload(keyPayload))
}

// Key format: KeyPrefixParticipant+"-"+Hash(PubKey+"-"+GroupId)
func (m *Participant) Key() []byte {
	keyPayload := hash256.FromBytes(JoinWithHyphen([][]byte{m.PubKey.Bytes(), m.GroupId.Bytes()}))
	return Key(KeyPrefixParticipant, KeyPayload(keyPayload))
}

// Key format: KeyPrefixGeneratedPublicKey+"-"+Hash(GenPubKey)
func (m *GeneratedPublicKey) Key() []byte {
	keyPayload := hash256.FromBytes(m.GenPubKey[:])
	return Key(KeyPrefixGeneratedPublicKey, KeyPayload(keyPayload))
}

func (m *GeneratedPublicKey) KeyFromHash(hash common.Hash) []byte {
	return Key(KeyPrefixGeneratedPublicKey, KeyPayload(hash))
}

func (m *StakeRequest) Key() []byte {
	return Key(KeyPrefixStakeRequest, KeyPayload(m.TxHash))
}

// Handy methods

func (m *Participant) ParticipantId() [32]byte {
	var indexByte []byte
	binary.BigEndian.PutUint64(indexByte, m.Index)

	var partiId [32]byte
	copy(partiId[:], m.GroupId[:])
	partiId[31] = indexByte[0]

	return partiId
}
