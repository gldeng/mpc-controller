package crypto

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding signing and verification

type Signer interface {
	Sign(message []byte) ([]byte, error)
}

type SignHasher interface {
	SignHash(hash []byte) ([]byte, error)
}

type Verifier interface {
	Verify(message, signature []byte) bool
}

type VerifyHasher interface {
	VerifyHash(hash, signature []byte) bool
}
