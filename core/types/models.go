package types

import (
	"bytes"
	"crypto/ecdsa"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/utils/address"
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
	return common.Bytes2Hex(m)
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
	return *address.PubkeyToAddresse(pubKey), nil
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

const (
	InitBit = "800000000000000000"
)

type ParticipantId [32]byte

func (m ParticipantId) Index() uint64 {
	return new(big.Int).SetBytes(m[31:32]).Uint64()
}

func (m ParticipantId) GroupId() [32]byte {
	var groupId [32]byte
	copy(groupId[:], m[:31])
	return groupId
}

func (m ParticipantId) Joined(indices *big.Int) bool {
	myIndex := m.Index() - 1
	initBit, _ := new(big.Int).SetString(InitBit, 16)
	myConfirm := new(big.Int).Rsh(initBit, uint(myIndex))
	and := new(big.Int).And(indices, myConfirm)
	zero := big.NewInt(0)
	return and.Cmp(zero) == 1
}

// Indices

type Indices big.Int

func (m *Indices) Indices() []uint {
	var joinedPartiIndices []uint
	for i := 0; i < 64; i++ {
		indices := (big.Int)(*m)
		initBit, _ := new(big.Int).SetString(InitBit, 16)
		myConfirm := new(big.Int).Rsh(initBit, uint(i))
		and := new(big.Int).And(&indices, myConfirm)
		zero := big.NewInt(0)
		if and.Cmp(zero) == 1 {
			joinedPartiIndices = append(joinedPartiIndices, uint(i+1))
		}
	}
	return joinedPartiIndices
}

func (m ParticipantId) Threshold() uint64 {
	return new(big.Int).SetBytes(m[30:31]).Uint64()
}

func (m ParticipantId) GroupSize() uint64 {
	return new(big.Int).SetBytes(m[29:30]).Uint64()
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
	TaskTypRecover
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
		return TaskTypRecover
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
	reqHash := new(big.Int).SetBytes(m[:])
	reqHash = new(big.Int).And(mask, reqHash)
	typ := new(big.Int).SetUint64(uint64(t))
	reqHashTyp := new(big.Int).Or(reqHash, typ)
	copy(m[:], reqHashTyp.Bytes())
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
	reqNoHash := hash256.FromBytes(new(big.Int).SetUint64(m.ReqNo).Bytes())
	reqHash := RequestHash(reqNoHash)
	reqHash.SetTaskType(TaskTypStake)
	return reqHash
}

// RecoverRequest

type RecoverRequest struct {
	TxID        ids.ID `json:"txID"`
	OutputIndex uint32 `json:"outputIndex"`

	*GeneratedPublicKey `json:"genPubKey"`
}

func (m *RecoverRequest) ReqHash() RequestHash {
	bs := new(big.Int).SetUint64(uint64(m.OutputIndex)).Bytes()
	reqHash := RequestHash(hash256.FromBytes(JoinWithHyphen([][]byte{m.TxID[:], bs})))
	reqHash.SetTaskType(TaskTypRecover)
	return reqHash
}

func Key(prefix []byte, payload [32]byte) []byte {
	return JoinWithHyphen([][]byte{prefix, payload[:]})
}

func JoinWithHyphen(s [][]byte) []byte {
	return bytes.Join(s, []byte("-"))
}
