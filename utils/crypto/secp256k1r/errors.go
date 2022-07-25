package secp256k1r

const (
	ErrMsgFailedToRecoveryPubKey = "failed to recovery public key"
	ErrMsgInvalidRecoveredPubKey = "invalid recovered public key"
)

// ----------error types ----------

type ErrTypPubKeyRecoveryFailure struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypPubKeyRecoveryFailure) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgFailedToRecoveryPubKey
	}
	return e.ErrMsg
}

func (e *ErrTypPubKeyRecoveryFailure) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypInvalidRecoveredPubKey struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypInvalidRecoveredPubKey) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgInvalidRecoveredPubKey
	}
	return e.ErrMsg
}

func (e *ErrTypInvalidRecoveredPubKey) Unwrap() error {
	return e.Cause
}
