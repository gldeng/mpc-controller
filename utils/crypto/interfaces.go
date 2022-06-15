package crypto

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding signing and verification

type Sign interface {
	Sign(message []byte) ([]byte, error)
}

type SignHash interface {
	SignHash(hash []byte) ([]byte, error)
}

type Verify interface {
	Verify(message, signature []byte) bool
}

type VerifyHash interface {
	VerifyHash(hash, signature []byte) bool
}
