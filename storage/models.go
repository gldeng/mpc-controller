package storage

import "time"

// Prefixes
const (
	prefixGroupInfo           = "groupInfo"
	prefixParticipantInfo     = "partyInfo"
	prefixGeneratedPubKeyInfo = "genPubKeyInfo"
	prefixKeygenRequestInfo   = "keygenReqInfo"
)

// GroupInfo stored key format: prefixGroupInfo-GroupIdHex
type GroupInfo struct {
	GroupIdHex     string
	PartPubKeyHexs []string
	Threshold      uint64
}

// ParticipantInfo stored key format: prefixParticipantInfo-PubKeyHashHex-GroupIdHex
type ParticipantInfo struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
	Index         uint64
}

// GeneratedPubKeyInfo stored key format: prefixGeneratedPubKeyInfo-PubKeyHashHex
type GeneratedPubKeyInfo struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
}

// KeygenRequestInfo stored key format: prefixKeygenRequestInfo-RequestIdHex
// todo: customize more agreeable time format
type KeygenRequestInfo struct {
	RequestIdHex     string
	GroupIdHex       string
	RequestAddedAt   time.Time
	PubKeyReportedAt time.Time
	PubKeyHashHex    string
}

// todo: add stake request info
