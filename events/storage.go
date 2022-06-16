package events

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning local storage

type GroupInfoStoredEvent struct {
	Key string
	Val GroupInfo
}

type ParticipantInfoStoredEvent struct {
	Key string
	Val ParticipantInfo
}

type GeneratedPubKeyInfoStoredEvent struct {
	Key string
	Val GeneratedPubKeyInfo
}

// Info types

// Prefixes
const (
	prefixGroupInfo           = "groupInfo"
	prefixParticipantInfo     = "partyInfo"
	prefixGeneratedPubKeyInfo = "genPubKeyInfo"
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

// GeneratedPubKeyInfo stored key format: prefixGeneratedPubKeyInfo-GenPubKeyHashHex
type GeneratedPubKeyInfo struct {
	GenPubKeyHashHex string
	GenPubKeyHex     string
	GroupIdHex       string
}
