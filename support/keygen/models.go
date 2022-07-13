package keygen

// Prefixes
const (
	prefixGeneratedPubKeyInfo = "genPubKeyInfo"
)

// GeneratedPubKeyInfo stored key format: prefixGeneratedPubKeyInfo-GenPubKeyHashHex
type GeneratedPubKeyInfo struct {
	GenPubKeyHashHex       string `json:"genPubKeyHashHex"`
	CompressedGenPubKeyHex string `json:"compressedGenPubKeyHex"`
	GroupIdHex             string `json:"groupIdHex"`
}
