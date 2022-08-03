package events

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning local storage

type GroupInfoStored struct {
	Key string
	Val GroupInfo
}

type ParticipantInfoStored struct {
	Key string
	Val ParticipantInfo
}

type GeneratedPubKeyInfoStored struct {
	Key string
	Val GeneratedPubKeyInfo
}

// Info types

// Prefixes
const (
	PrefixGroupInfo           = "groupInfo"
	PrefixParticipantInfo     = "partyInfo"
	PrefixGeneratedPubKeyInfo = "genPubKeyInfo"
)

// GroupInfo stored key format: PrefixGroupInfo-GroupIdHex
type GroupInfo struct {
	GroupIdHex     string
	PartPubKeyHexs []string
	Threshold      uint64
}

// ParticipantInfo stored key format: PrefixParticipantInfo-PubKeyHashHex-GroupIdHex
type ParticipantInfo struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
	Index         uint64
}

// GeneratedPubKeyInfo stored key format: PrefixGeneratedPubKeyInfo-GenPubKeyHashHex
type GeneratedPubKeyInfo struct {
	GenPubKeyHashHex       string
	CompressedGenPubKeyHex string
	GroupIdHex             string
}
