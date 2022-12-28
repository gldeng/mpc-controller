package services

import (
	"github.com/avalido/mpc-controller/utils/crypto"
)

// Types for network request input and output

type KeygenInput struct {
	RequestId       string   `json:"request_id"`
	ParticipantKeys []string `json:"public_keys"`
	Threshold       int      `json:"t"`
}

type SignInput struct {
	RequestId       string   `json:"request_id"`
	PublicKey       string   `json:"public_key"`
	ParticipantKeys []string `json:"participant_public_keys"`
	Hash            string   `json:"message"`
}

type ResultInput struct {
	RequestId string `path:"reqId"`
}

type ResultOutput struct {
	RequestId     string        `json:"request_id"`
	Result        string        `json:"result"`
	RequestType   RequestType   `json:"request_type"`
	RequestStatus RequestStatus `json:"request_status"`
}

// Types for internal status recordings

type RequestType string

const (
	TypeKeygen RequestType = "KEYGEN"
	TypeSign   RequestType = "SIGN"
)

type RequestStatus string

const (
	StatusReceived   RequestStatus = "RECEIVED"
	StatusProcessing RequestStatus = "PROCESSING"
	StatusDone       RequestStatus = "DONE"

	StatusOfflineStageDone RequestStatus = "OFFLINE_STAGE_DONE" // for sign request only
	StatusError            RequestStatus = "ERROR"              // for sign request only, what about keygen request?
)

type KeygenRequestModel struct {
	input   *KeygenInput
	reqType RequestType

	hits   int
	status RequestStatus

	signer crypto.Signer_

	result string // public key hex string
}

type SignRequestModel struct {
	input   *SignInput
	reqType RequestType

	hits   int
	status RequestStatus

	result string // hex signature string
}
