package addDelegator

const (
	ErrMsgSignatureAlreadySet           = "signature already set"
	ErrMsgPubKeyCannotRecover           = "public key cannot recover"
	ErrMsgSignatureInvalid              = "invalid signature"
	ErrMsgMissingCredential             = "missing credential"
	ErrMsgInvalidUint64                 = "invalid uint64"
	ErrMsgFailedToGetNonce              = "failed to get nonce"
	ErrMsgFailedToConvertAmount         = "failed to convert amount"
	ErrMsgFailedToBuildTx               = "failed to build tx"
	ErrMsgFailedToBuildAndSignTx        = "failed to build and sign tx"
	ErrMsgFailedToGetSignatureAndSendTx = "failed to get signature and send tx"
	ErrMsgFailedToCheckStatus           = "failed to check status"
	ErrMsgFailedToGetTxHash             = "failed to get tx hash"
	ErrMsgFailedToCreateSignRequest     = "failed to create sign request"
	ErrMsgFailedToSendSignRequest       = "failed to post signing request"
	ErrMsgFailedToCheckSignRequest      = "failed to check signing result"
	DebugMsgSignRequestNotDone          = "sign request is still pending"
	ErrMsgFailedToValidateCredential    = "failed to validate cred"
	ErrMsgFailedToPrepareSignedTx       = "failed to get signed tx"
	ErrMsgFailedToIssueTx               = "failed to issue tx"
)
