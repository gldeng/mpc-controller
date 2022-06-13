package events

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning local storage

type GroupInfoStoredEvent struct {
	GroupIdHex     string
	PartPubKeyHexs []string
	Threshold      uint64
}

type ParticipantInfoStoredEvent struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
	Index         uint64
}

type GeneratedPubKeyInfoStoredEvent struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
}
