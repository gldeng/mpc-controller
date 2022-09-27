package events

import "github.com/avalido/mpc-controller/utils/bytes"

const (
	ReqTypKeygen   ReqType = "KEYGEN"
	ReqTypSignSign ReqType = "SIGN"
)

const (
	ReqStatusReceived   ReqStatus = "RECEIVED"
	ReqStatusProcessing ReqStatus = "PROCESSING"
	ReqStatusDone       ReqStatus = "DONE"
)

const (
	ReqIDPrefixKeygen ReqIDPrefix = "KEYGEN-"

	ReqIDPrefixSignStake             ReqIDPrefix = "SIGN-STAKE-"
	ReqIDPrefixSignStakeExport       ReqIDPrefix = "SIGN-STAKE-EXPORT-"
	ReqIDPrefixSignStakeImport       ReqIDPrefix = "SIGN-STAKE-IMPORT-"
	ReqIDPrefixSignStakeAddDelegator ReqIDPrefix = "SIGN-STAKE-ADD-DELEGATOR-"

	ReqIDPrefixSignPrincipal       ReqIDPrefix = "SIGN-PRINCIPAL-"
	ReqIDPrefixSignPrincipalExport ReqIDPrefix = "SIGN-PRINCIPAL-EXPORT-"
	ReqIDPrefixSignPrincipalImport ReqIDPrefix = "SIGN-PRINCIPAL-IMPORT-"

	ReqIDPrefixSignReward       ReqIDPrefix = "SIGN-REWARD-"
	ReqIDPrefixSignRewardExport ReqIDPrefix = "SIGN-REWARD-EXPORT-"
	ReqIDPrefixSignRewardImport ReqIDPrefix = "SIGN-REWARD-IMPORT-"
)

type ReqType string
type ReqStatus string
type ReqIDPrefix string

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
	Result *Signature
}
