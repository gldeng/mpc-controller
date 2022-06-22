package participant

// Prefixes
const (
	prefixGroupInfo       = "groupInfo"
	prefixParticipantInfo = "partyInfo"
)

// GroupInfo stored key format: prefixGroupInfo-GroupIdHex
type GroupInfo struct {
	GroupIdHex     string   `json:"groupIdHex"`
	PartPubKeyHexs []string `json:"partPubKeyHexs"`
	Threshold      uint64   `json:"threshold"`
}

// ParticipantInfo stored key format: prefixParticipantInfo-PubKeyHashHex-GroupIdHex
type ParticipantInfo struct {
	PubKeyHashHex string `json:"pubKeyHashHex"`
	PubKeyHex     string `json:"pubKeyHex"`
	GroupIdHex    string `json:"groupIdHex"`
	Index         uint64 `json:"index"`
}
