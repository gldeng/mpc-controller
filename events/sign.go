package events

import "github.com/avalido/mpc-controller/utils/bytes"

const (
	ReqTypKeygen   ReqType = "KEYGEN"
	ReqTypSignSign ReqType = "SIGN"
)

const (
	ReqStatusSubmitted  SignStatus = "SUBMITTED"
	ReqStatusReceived   SignStatus = "RECEIVED"
	ReqStatusProcessing SignStatus = "PROCESSING"
	ReqStatusDone       SignStatus = "DONE"
)

const (
	SignIDPrefixStakeExport       IDPrefix = "STAKE-EXPORT-"
	SignIDPrefixStakeImport       IDPrefix = "STAKE-IMPORT-"
	SignIDPrefixStakeAddDelegator IDPrefix = "STAKE-ADD-DELEGATOR-"

	SignIDPrefixSignPrincipalExport IDPrefix = "RECOVER-PRINCIPAL-EXPORT-"
	SignIDPrefixSignPrincipalImport IDPrefix = "RECOVER-PRINCIPAL-IMPORT-"

	SignIDPrefixSignRewardExport IDPrefix = "RECOVER-REWARD-EXPORT-"
	SignIDPrefixSignRewardImport IDPrefix = "RECOVER-REWARD-IMPORT-"
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
type SignStatus string
type IDPrefix string
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
