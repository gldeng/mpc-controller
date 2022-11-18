package types

const (
	TypKeygen   Type = "KEYGEN"
	TypSignSign Type = "SIGN"
)

const (
	StatusReceived   Status = "RECEIVED"
	StatusProcessing Status = "PROCESSING"
	StatusDone       Status = "DONE"
)

type Type string
type Status string

type KeygenRequest struct {
	ReqID                  string   `json:"request_id"`
	CompressedPartiPubKeys []string `json:"public_keys"`
	Threshold              uint64   `json:"t"`
}

type SignRequest struct {
	ReqID                  string   `json:"request_id"`
	Hash                   string   `json:"message"`
	CompressedGenPubKeyHex string   `json:"public_key"`              // TODO: add format check
	CompressedPartiPubKeys []string `json:"participant_public_keys"` // TODO: add format check
}

type Result struct {
	ReqID  string `json:"request_id"`
	Result string `json:"result"`
	Type   Type   `json:"request_type"`
	Status Status `json:"request_status"`
}
