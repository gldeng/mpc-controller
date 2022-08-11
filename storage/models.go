package storage

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	ids2 "github.com/avalido/mpc-controller/utils/ids"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
)

// todo: use struct tag and reflect to deal with key.

// Key prefixes
var (
	KeyPrefixGroup              = []byte("group")
	KeyPrefixParticipant        = []byte("parti")
	KeyPrefixGeneratedPublicKey = []byte("genPubKey")

	KeyPrefixJoinRequest = []byte("JoinReq")
)

// PartiPubKey

type PubKey []byte // uncompressed

func (m PubKey) String() string {
	return m.PubKeyHex()
}

func (m PubKey) PubKeyHex() string {
	return bytes.BytesToHex(m)
}

func (m PubKey) CChainAddress() (common.Address, error) {
	cmp, err := m.CompressPubKey()
	if err != nil {
		return *new(common.Address), errors.WithStack(err)
	}
	pubKey, err := crypto.UnmarshalPubkeyBytes(cmp)
	if err != nil {
		return *new(common.Address), errors.WithStack(err)
	}
	return *addrs.PubkeyToAddresse(pubKey), nil
}

func (m PubKey) PChainAddress() (ids.ShortID, error) {
	cmp, err := m.CompressPubKey()
	if err != nil {
		return *new(ids.ShortID), nil
	}
	id, err := ids2.ShortIDFromPubKeyBytes(cmp)
	if err != nil {
		return *new(ids.ShortID), nil
	}
	return *id, nil
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

func (m PubKey) ECDSAPubKey() (*ecdsa.PublicKey, error) {
	pk, err := crypto.UnmarshalPubkeyBytes(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return pk, nil
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
	return new(big.Int).SetBytes(m[31:32]).Uint64()
}

func (m ParticipantId) Joined(indices *big.Int) bool {
	return indices.Bit(int(m.Index())-1) > 0
}

func (m ParticipantId) Threshold() uint64 {
	return new(big.Int).SetBytes(m[30:31]).Uint64()
}

func (m ParticipantId) GroupSize() uint64 {
	return new(big.Int).SetBytes(m[29:30]).Uint64()
}

// Group

type Group struct {
	ID    common.Hash `json:"id"`
	Group PubKeys     `json:"group"`
}

func (m *Group) Key() []byte { // Key format: KeyPrefixGroup+"-"+ID
	keyPayload := m.ID
	return Key(KeyPrefixGroup, keyPayload)
}

func (m *Group) GroupSize() uint64 {
	return new(big.Int).SetBytes(m.ID[29:30]).Uint64()
}

func (m *Group) Threshold() uint64 {
	return new(big.Int).SetBytes(m.ID[30:31]).Uint64()
}

// Participant

type Participant struct {
	PubKey  common.Hash `json:"pubKey"`
	GroupId common.Hash `json:"groupId"`
	Index   uint64      `json:"index"`
}

func (m *Participant) Key() []byte { // Key format: KeyPrefixParticipant+"-"+Hash(PartiPubKey+"-"+GroupId)
	keyPayload := hash256.FromBytes(JoinWithHyphen([][]byte{m.PubKey.Bytes(), m.GroupId.Bytes()}))
	return Key(KeyPrefixParticipant, keyPayload)
}

func (m *Participant) ParticipantId() ParticipantId {
	groupIdBig := new(big.Int).SetBytes(m.GroupId[:])
	indexBig := new(big.Int).SetUint64(m.Index)
	partiIdBig := new(big.Int).Or(groupIdBig, indexBig)
	var partiId [32]byte
	copy(partiId[:], partiIdBig.Bytes())
	return partiId
}

// GeneratedPublicKey

type GeneratedPublicKey struct {
	GenPubKey PubKey      `json:"genPubKey"`
	GroupId   common.Hash `json:"groupId"`
}

func (m *GeneratedPublicKey) Key() []byte { // Key format: KeyPrefixGeneratedPublicKey+"-"+Hash(GenPubKey)
	keyPayload := hash256.FromBytes(m.GenPubKey[:])
	return Key(KeyPrefixGeneratedPublicKey, keyPayload)
}

func (m *GeneratedPublicKey) KeyFromHash(hash common.Hash) []byte {
	return Key(KeyPrefixGeneratedPublicKey, hash)
}

// RequestHash

const (
	TaskTypUnknown TaskType = iota
	TaskTypStake
	TaskTypReturn
)

const (
	Init32ByteMask = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00"
)

type TaskType uint8

type RequestHash [32]byte

func (m *RequestHash) TaskType() TaskType {
	n := new(big.Int).SetBytes(m[31:32]).Uint64()
	switch {
	case n == 1:
		return TaskTypStake
	case n == 2:
		return TaskTypReturn
	default:
		return TaskTypUnknown
	}
}

func (m *RequestHash) IsTaskType(t TaskType) bool {
	reqTyp := new(big.Int).SetBytes(m[31:32]).Uint64()
	return uint64(t) == reqTyp
}

func (m *RequestHash) SetTaskType(t TaskType) {
	mask, _ := new(big.Int).SetString(Init32ByteMask, 16)
	fmt.Printf("mask: %v, hex: %v\n", mask.String(), bytes.BytesToHex(mask.Bytes()))
	reqHash := new(big.Int).SetBytes(m[:])
	fmt.Printf("reqHash: %v, hex:%v\n", reqHash.String(), bytes.BytesToHex(reqHash.Bytes()))

	reqHash = new(big.Int).And(mask, reqHash)
	fmt.Printf("reqHashAnd: %v, hex:%v\n", reqHash.String(), bytes.BytesToHex(reqHash.Bytes()))

	typ := new(big.Int).SetUint64(uint64(t))
	reqHashTyp := new(big.Int).Or(reqHash, typ)
	fmt.Printf("reqHashTyp: %v, hex:%v\n", reqHashTyp.String(), bytes.BytesToHex(reqHashTyp.Bytes()))

	copy(m[:], reqHashTyp.Bytes())
	fmt.Printf("final reqHash hex:%v\n", m.String())
}

func (m *RequestHash) String() string {
	return common.Hash(*m).String()
}

// JoinRequest

type JoinRequest struct {
	ReqHash RequestHash   `json:"reqHash"`
	PartiId ParticipantId `json:"partiId"`
	Args    interface{}   `json:"args"`
}

func (m *JoinRequest) Key() []byte { // Key format: KeyPrefixJoinRequest+"-"+ReqHash
	return Key(KeyPrefixJoinRequest, m.ReqHash)
}

// StakeRequest

type StakeRequest struct {
	ReqNo     uint64      `json:"reqNo"`
	TxHash    common.Hash `json:"txHash"`
	NodeID    string      `json:"nodeID"`
	Amount    string      `json:"amount"`
	StartTime int64       `json:"startTime"`
	EndTime   int64       `json:"endTime"`

	*GeneratedPublicKey `json:"genPubKey"`
}

func (m *StakeRequest) ReqHash() RequestHash {
	reqHash := RequestHash(m.TxHash)
	reqHash.SetTaskType(TaskTypStake)
	return reqHash
}

// ExportUTXORequest

type ExportUTXORequest struct {
	TxID        ids.ID `json:"txID"`
	OutputIndex uint32 `json:"outputIndex"`

	*GeneratedPublicKey `json:"genPubKey"`
}

func (m *ExportUTXORequest) ReqHash() RequestHash {
	bs := new(big.Int).SetUint64(uint64(m.OutputIndex)).Bytes()
	reqHash := RequestHash(hash256.FromBytes(JoinWithHyphen([][]byte{m.TxID[:], bs})))
	reqHash.SetTaskType(TaskTypReturn)
	return reqHash
}
