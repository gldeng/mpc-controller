package support

import "github.com/ethereum/go-ethereum/common"

// Prefixes
var (
	KeyPrefixGroup              = []byte("group-")
	KeyPrefixParticipant        = []byte("parti-")
	KeyPrefixGeneratedPublicKey = []byte("genPubKey-")
)

const (
	PubKeyLength = 64
)

type PubKey [PubKeyLength]byte

// Group stored key format: KeyPrefixGroup+ID
type Group struct {
	ID    common.Hash `json:"id"`
	Group []PubKey    `json:"group"`
}

// Participant stored key format: KeyPrefixParticipant+PubKey+GroupId
type Participant struct {
	PubKey  common.Hash `json:"pubKey"`
	GroupId common.Hash `json:"groupId"`
	Index   uint64      `json:"index"`
}

// GeneratedPublicKey stored key format: KeyPrefixGeneratedPublicKey+GenPubKey
type GeneratedPublicKey struct {
	GenPubKey PubKey      `json:"genPubKey"`
	GroupId   common.Hash `json:"groupId"`
}
