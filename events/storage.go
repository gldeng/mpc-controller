package events

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning local storage

type GroupInfoStoredEvent struct {
	Key string
	Val GroupInfo
}

type ParticipantInfoStoredEvent struct {
	Key string
	Val ParticipantInfoStored
}

type GeneratedPubKeyInfoStoredEvent struct {
	Key string
	Val GeneratedPubKeyInfo
}

// Info types

type GroupInfo struct {
	GroupIdHex     string
	PartPubKeyHexs []string
	Threshold      uint64
}

type ParticipantInfoStored struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
	Index         uint64
}

type GeneratedPubKeyInfo struct {
	PubKeyHashHex string
	PubKeyHex     string
	GroupIdHex    string
}
