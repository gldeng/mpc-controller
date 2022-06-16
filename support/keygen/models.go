package keygen

// Prefixes
const (
	prefixGeneratedPubKeyInfo = "genPubKeyInfo"
)

// GeneratedPubKeyInfo stored key format: prefixGeneratedPubKeyInfo-GenPubKeyHashHex
type GeneratedPubKeyInfo struct {
	GenPubKeyHashHex string
	GenPubKeyHex     string
	GroupIdHex       string
}
