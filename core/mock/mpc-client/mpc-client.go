package mpc_client

import (
	"context"
	"errors"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"math/rand"
	"sync"
	"time"
)

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

type RequestModel struct {
	reqID   string
	_type   RequestType
	status  RequestStatus
	payload interface{}

	err    error
	result string // for key sign request, it is public key string; for sign request, it is signature string.

	signer crypto.Signer
}

// todo: use distributed cache, or persistent db
var requestsCache = &sync.Map{}

type MPCClient interface {
	Keygen(ctx context.Context, keygenReq *core.KeygenRequest) error
	Sign(ctx context.Context, signReq *core.SignRequest) error
	Result(ctx context.Context, reqID string) (*core.Result, error)
}

type MpcClientMock struct {
	parties   int
	threshold int
}

func New(parties, threshold int) core.MPCClient {
	return &MpcClientMock{parties, threshold}
}

// todo: reuse key for a group?
// todo: to return response  with error?
func (m *MpcClientMock) Keygen(ctx context.Context, keygenReq *core.KeygenRequest) error {
	go func() {
		// simulate time elapse before process keygen
		time.Sleep(time.Second)
		logger.Debug("Keygen request PROCESSING.", logger.Field{"requestId", keygenReq.RequestId})

		value, ok := requestsCache.Load(keygenReq.RequestId)
		if ok { // todo: deal with !ok case
			reqMod := value.(*RequestModel)
			reqMod.status = StatusProcessing
			requestsCache.Store(keygenReq.RequestId, reqMod)
		}

		// simulate processing keygen, which may take some time, including waiting for all parties fully call
		for !m.isAllPartiesCalledKeygen() {
			time.Sleep(3)
		}
		signer, err := crypto.NewSECP256K1RSigner()
		if err != nil {
		} // todo: deal with error,
		value, ok = requestsCache.Load(keygenReq.RequestId)
		if ok { // todo: deal with !ok case
			reqMod := value.(*RequestModel)
			reqMod.signer = signer
			reqMod.result = string(signer.PublicKey().Bytes())
			reqMod.status = StatusDone
			requestsCache.Store(keygenReq.RequestId, reqMod)
			logger.Debug("Keygen request DONE",
				logger.Field{"requestID", keygenReq.RequestId},
				logger.Field{"privateKey", string(signer.PrivateKey().Bytes())},
				logger.Field{"publicKey", reqMod.result},
				logger.Field{"addressHex", signer.Address().Hex()})
		}
	}()

	requestsCache.Store(keygenReq.RequestId, &RequestModel{
		reqID:   keygenReq.RequestId,
		_type:   TypeKeygen,
		status:  StatusReceived,
		payload: keygenReq})
	logger.Debug("Keygen request RECEIVED.", logger.Field{"requestId", keygenReq.RequestId})
	return nil
}

// todo: to return response with error?
func (m *MpcClientMock) Sign(ctx context.Context, signReq *core.SignRequest) error {
	go func() {
		// simulate time elapse before process sign
		time.Sleep(time.Second)
		logger.Debug("Sign request PROCESSING.", logger.Field{"requestId", signReq.RequestId})

		value, ok := requestsCache.Load(signReq.RequestId)
		if ok { // todo: deal with !ok case
			reqMod := value.(*RequestModel)
			reqMod.status = StatusProcessing
			requestsCache.Store(signReq.RequestId, reqMod)
		}

		// simulate processing sign, which may take some time, including waiting for enough parties call
		for !m.isEnoughPartiesCalledSign() {
			time.Sleep(3)
		}

		value, ok = requestsCache.Load(signReq.RequestId)
		if ok { // todo: deal with !ok case
			reqMod := value.(*RequestModel)
			signer := reqMod.signer
			reqPayload := reqMod.payload.(*core.SignRequest)
			sig, err := signer.SignHash([]byte(reqPayload.Hash)) // Note whether it is correct input
			if err != nil {
				// todo: deal with this err
			}
			reqMod.result = string(sig)
			reqMod.status = StatusDone
			requestsCache.Store(signReq.RequestId, reqMod)
			logger.Debug("sign request DONE",
				logger.Field{"requestID", signReq.RequestId},
				logger.Field{"signature", reqMod.result})
		}
	}()

	requestsCache.Store(signReq.RequestId, &RequestModel{
		reqID:   signReq.RequestId,
		_type:   TypeSign,
		status:  StatusReceived,
		payload: signReq})
	return nil
}
func (m *MpcClientMock) Result(ctx context.Context, reqID string) (*core.Result, error) {
	req, ok := requestsCache.Load(reqID)
	if !ok {
		return nil, errors.New(reqID + " not found")
	}
	reqMod := req.(*RequestModel)
	return &core.Result{
		RequestId:     reqMod.reqID,
		RequestType:   string(reqMod._type),
		RequestStatus: string(reqMod.status),
		Result:        reqMod.result,
	}, nil
}

func (m *MpcClientMock) isAllPartiesCalledKeygen() bool {
	rand.Seed(time.Now().UnixNano())
	randNun := rand.Intn(m.parties + 1)
	if randNun == m.parties {
		return true
	}
	return false
}

func (m *MpcClientMock) isEnoughPartiesCalledSign() bool {
	rand.Seed(time.Now().UnixNano())
	randNun := rand.Intn(m.threshold + 2)
	if randNun == m.threshold+1 {
		return true
	}
	return false
}
