package events

import "github.com/avalido/mpc-controller/utils/bytes"

const (
	ReqTypKeygen   ReqType = "KEYGEN"
	ReqTypSignSign ReqType = "SIGN"
)

const (
	ReqStatusSubmitted  ReqStatus = "SUBMITTED"
	ReqStatusReceived   ReqStatus = "RECEIVED"
	ReqStatusProcessing ReqStatus = "PROCESSING"
	ReqStatusDone       ReqStatus = "DONE"
)

const (
	SignIDPrefixStakeExport       SignIDPrefix = "SIGN-STAKE-EXPORT-"
	SignIDPrefixStakeImport       SignIDPrefix = "SIGN-STAKE-IMPORT-"
	SignIDPrefixStakeAddDelegator SignIDPrefix = "SIGN-STAKE-ADD-DELEGATOR-"

	SignIDPrefixSignPrincipalExport SignIDPrefix = "SIGN-PRINCIPAL-EXPORT-"
	SignIDPrefixSignPrincipalImport SignIDPrefix = "SIGN-PRINCIPAL-IMPORT-"

	SignIDPrefixSignRewardExport SignIDPrefix = "SIGN-REWARD-EXPORT-"
	SignIDPrefixSignRewardImport SignIDPrefix = "SIGN-REWARD-IMPORT-"
)

const (
	SignKindStakeExport SignKind = iota
	SignKindStakeImport
	SignKindStakeAddDelegator

	SignKindPrincipalExport
	SignKindPrincipalImport

	SignKindRewardExport
	SignKindRewardImport
)

type ReqType string
type ReqStatus string
type SignIDPrefix string
type SignKind int

type Signature [65]byte

func (s *Signature) FromHex(hex string) *Signature {
	*s = bytes.BytesTo65Bytes(bytes.HexToBytes(hex))
	return s
}

func (s *Signature) String() string {
	return bytes.Bytes65ToHex(*s)
}

// todo: KeygenDone

type SignDone struct {
	ReqID  string
	Kind   SignKind
	Result *Signature
}
