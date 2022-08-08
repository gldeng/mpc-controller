package storage

import (
	"encoding/binary"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// todo: use struct tag and reflect to deal with key.

// Key prefixes
var (
	KeyPrefixGroup              KeyPrefix = []byte("group")
	KeyPrefixParticipant        KeyPrefix = []byte("parti")
	KeyPrefixGeneratedPublicKey KeyPrefix = []byte("genPubKey")

	KeyPrefixJoinRequest  KeyPrefix = []byte("JoinReq")
	KeyPrefixStakeRequest KeyPrefix = []byte("stakeReq")
)

type ReqHash [32]byte

// PubKey

type PubKey []byte // uncompressed

func (m PubKey) CChainAddress() (common.Address, error) {
	pubKey, err := crypto.UnmarshalPubkeyBytes(m)
	if err != nil {
		return *new(common.Address), errors.WithStack(err)
	}
	return *addrs.PubkeyToAddresse(pubKey), nil
}

func (m PubKey) CompressPubKey() ([]byte, error) {
	normed, err := crypto.NormalizePubKeyBytes(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return normed, nil
}

func (m PubKey) CompressPubKeyHex() (string, error) {
	normed, err := crypto.NormalizePubKey(common.Bytes2Hex(m))
	if err != nil {
		return "", errors.WithStack(err)
	}
	return *normed, nil
}

// PubKeys

type PubKeys []PubKey

func (m PubKeys) CompressPubKeys() ([][]byte, error) {
	var cmps [][]byte
	for _, pubKey := range m {
		cmp, err := pubKey.CompressPubKey()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		cmps = append(cmps, cmp)
	}
	return cmps, nil
}

func (m PubKeys) CompressPubKeyHexs() ([]string, error) {
	var cmps []string
	for _, pubKey := range m {
		cmp, err := pubKey.CompressPubKeyHex()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		cmps = append(cmps, cmp)
	}
	return cmps, nil
}

// ParticipantId

type ParticipantId [32]byte

func (m ParticipantId) Index() uint64 {
	return binary.BigEndian.Uint64(m[31:])
}

func (m ParticipantId) Threshold() uint64 {
	return binary.BigEndian.Uint64(m[30:31])
}

func (m ParticipantId) GroupSize() uint64 {
	return binary.BigEndian.Uint64(m[29:30])
}

// Group

type Group struct {
	ID    common.Hash `json:"id"`
	Group PubKeys     `json:"group"`
}

func (m *Group) Key() []byte { // Key format: KeyPrefixGroup+"-"+ID
	keyPayload := m.ID
	return Key(KeyPrefixGroup, KeyPayload(keyPayload))
}

// Participant

type Participant struct {
	PubKey  common.Hash `json:"pubKey"`
	GroupId common.Hash `json:"groupId"`
	Index   uint64      `json:"index"`
}

func (m *Participant) Key() []byte { // Key format: KeyPrefixParticipant+"-"+Hash(PubKey+"-"+GroupId)
	keyPayload := hash256.FromBytes(JoinWithHyphen([][]byte{m.PubKey.Bytes(), m.GroupId.Bytes()}))
	return Key(KeyPrefixParticipant, KeyPayload(keyPayload))
}

func (m *Participant) ParticipantId() ParticipantId {
	var indexByte []byte
	binary.BigEndian.PutUint64(indexByte, m.Index)

	var partiId [32]byte
	copy(partiId[:], m.GroupId[:])
	partiId[31] = indexByte[0]

	return partiId
}

// GeneratedPublicKey

type GeneratedPublicKey struct {
	GenPubKey PubKey      `json:"genPubKey"`
	GroupId   common.Hash `json:"groupId"`
}

func (m *GeneratedPublicKey) Key() []byte { // Key format: KeyPrefixGeneratedPublicKey+"-"+Hash(GenPubKey)
	keyPayload := hash256.FromBytes(m.GenPubKey[:])
	return Key(KeyPrefixGeneratedPublicKey, KeyPayload(keyPayload))
}

func (m *GeneratedPublicKey) KeyFromHash(hash common.Hash) []byte {
	return Key(KeyPrefixGeneratedPublicKey, KeyPayload(hash))
}

// JoinRequest

type JoinRequest struct {
	ReqHash ReqHash       `json:"reqHash"`
	PartiId ParticipantId `json:"partiId"`
	Args    interface{}   `json:"args"`
}

func (m *JoinRequest) Key() []byte { // Key format: KeyPrefixJoinRequest+"-"+ReqHash
	return Key(KeyPrefixJoinRequest, KeyPayload(m.ReqHash))
}

// StakeRequest

type StakeRequest struct {
	ReqNo     uint64      `json:"reqNo"`
	TxHash    common.Hash `json:"txHash"`
	GenPubKey common.Hash `json:"genPubKey"`
	NodeID    string      `json:"nodeID"`
	Amount    string      `json:"amount"`
	StartTime int64       `json:"startTime"`
	EndTime   int64       `json:"endTime"`
}

func (m *StakeRequest) Key() []byte { // Key format: KeyPrefixStakeRequest+"-"+TxHash
	return Key(KeyPrefixStakeRequest, KeyPayload(m.TxHash))
}
