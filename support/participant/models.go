package participant

// Prefixes
const (
	prefixGroupInfo       = "groupInfo"
	prefixParticipantInfo = "partyInfo"
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
