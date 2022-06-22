package keygen

// Prefixes
const (
	prefixGeneratedPubKeyInfo = "genPubKeyInfo"
)

// GeneratedPubKeyInfo stored key format: prefixGeneratedPubKeyInfo-GenPubKeyHashHex
type GeneratedPubKeyInfo struct {
	GenPubKeyHashHex string `json:"genPubKeyHashHex"`
	GenPubKeyHex     string `json:"genPubKeyHex"`
	GroupIdHex       string `json:"groupIdHex"`
}
