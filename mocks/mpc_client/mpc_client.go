package mpc_client

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
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

type keyGenRequestModel struct {
	status    RequestStatus
	payload   *core.KeygenRequest
	pubKeyHex string
	signer    crypto.Signer
}

type signRequestModel struct {
	status    RequestStatus
	payload   *core.SignRequest
	signature string
}

var keyGenRequestsCache = &sync.Map{}
var signRequestCache = &sync.Map{}

var pendingKeygenChan = make(chan string)
var pendingSignChan = make(chan string)

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
	m := &MpcClientMock{parties, threshold}
	go func() {
		for {
			select {
			case reqId := <-pendingKeygenChan:
				err := m.keygen(reqId)
				if err != nil {
					logger.Error("Failed to generate key",
						logger.Field{"requestId", reqId},
						logger.Field{"error", err})
				}
			case reqId := <-pendingSignChan:
				err := m.sign(reqId)
				if err != nil {
					logger.Error("Failed to sign message",
						logger.Field{"requestId", reqId},
						logger.Field{"error", err})
				}
			}
		}
	}()
	return m
}

func (m *MpcClientMock) Keygen(ctx context.Context, keygenReq *core.KeygenRequest) error {
	reqId := keygenReq.RequestId
	_, ok := keyGenRequestsCache.Load(reqId)
	if ok {
		return errors.Errorf("request id %q had been received, don't repeat call Keygen interface", reqId)
	}

	keyGenRequestsCache.Store(keygenReq.RequestId, &keyGenRequestModel{
		status:  StatusReceived,
		payload: keygenReq,
	})
	logger.Debug("Keygen request RECEIVED.", logger.Field{"requestId", reqId})

	pendingKeygenChan <- reqId
	return nil
}

func (m *MpcClientMock) Sign(ctx context.Context, signReq *core.SignRequest) error {
	reqId := signReq.RequestId
	_, ok := signRequestCache.Load(reqId)
	if ok {
		return errors.Errorf("request id %q had been received, don't repeat call Sign interface", reqId)
	}

	value, ok := keyGenRequestsCache.Load(reqId)
	if !ok {
		return errors.Errorf("you haven't call Keygen interface for request id %q, thus there's no signer to sign", reqId)
	}

	keyGenReqMod := value.(*keyGenRequestModel)
	if keyGenReqMod.status != StatusDone || keyGenReqMod.signer == nil {
		return errors.Errorf("request id %q is in keygen process and you should wait for a signer to be available.", reqId)
	}

	signRequestCache.Store(reqId, &signRequestModel{
		status:  StatusReceived,
		payload: signReq})

	pendingSignChan <- reqId
	return nil
}
func (m *MpcClientMock) Result(ctx context.Context, reqID string) (*core.Result, error) {
	value, ok := signRequestCache.Load(reqID)
	if ok {
		signReqMod := value.(*signRequestModel)
		return &core.Result{
			RequestId:     reqID,
			RequestType:   string(TypeSign),
			RequestStatus: string(signReqMod.status),
			Result:        signReqMod.signature,
		}, nil
	}

	value, ok = keyGenRequestsCache.Load(reqID)
	if ok {
		keygenMod := value.(*keyGenRequestModel)
		return &core.Result{
			RequestId:     reqID,
			RequestType:   string(TypeKeygen),
			RequestStatus: string(keygenMod.status),
			Result:        keygenMod.pubKeyHex,
		}, nil
	}

	return nil, errors.Errorf("not found result concerning request id %q", reqID)
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

func (m *MpcClientMock) keygen(reqId string) error {
	// simulate time elapse before process keygen
	time.Sleep(time.Second)
	logger.Debug("Keygen request PROCESSING.", logger.Field{"requestId", reqId})

	value, ok := keyGenRequestsCache.Load(reqId)
	if ok {
		reqMod := value.(*keyGenRequestModel)
		if reqMod.status == StatusDone && reqMod.signer != nil {
			return nil
		}
		reqMod.status = StatusProcessing
		keyGenRequestsCache.Store(reqId, reqMod)
	} else {
		return errors.Errorf("request id %q hasn't been received for keygen yet.", reqId)
	}

	// simulate processing keygen, which may take some time, including waiting for all parties fully call
	for !m.isAllPartiesCalledKeygen() {
		time.Sleep(3)
	}
	signer, err := crypto.NewSECP256K1RSigner()
	if err != nil {
		return errors.WithStack(err)
	}
	value, _ = keyGenRequestsCache.Load(reqId)
	reqMod := value.(*keyGenRequestModel)
	reqMod.signer = signer
	reqMod.pubKeyHex = common.Bytes2Hex(signer.PublicKey().Bytes())
	reqMod.status = StatusDone
	keyGenRequestsCache.Store(reqId, reqMod)
	logger.Debug("Keygen request DONE",
		logger.Field{"requestID", reqId},
		logger.Field{"privateKeyHex", common.Bytes2Hex(signer.PrivateKey().Bytes())},
		logger.Field{"publicKeyHex", reqMod.pubKeyHex},
		logger.Field{"addressHex", signer.Address().Hex()})
	return nil
}

func (m *MpcClientMock) sign(reqId string) error {
	// simulate time elapse before process sign
	time.Sleep(time.Second)
	logger.Debug("Sign request PROCESSING.", logger.Field{"requestId", reqId})

	value, ok := signRequestCache.Load(reqId)
	if ok {
		signReqMod := value.(*signRequestModel)
		if signReqMod.status == StatusDone && signReqMod.signature != "" {
			return errors.Errorf("request id %q hasn't signed before, don't repeat request sign", reqId)
		}

		value, ok := keyGenRequestsCache.Load(reqId)
		if !ok {
			return errors.Errorf("your haven't call keygen for request id %q yet, so there's no signer available", reqId)
		}
		keygenMod := value.(*keyGenRequestModel)
		signer := keygenMod.signer
		if signer == nil {
			return errors.Errorf("request id %q has been received but it's signer/key hasn't been generated", reqId)
		}

		signReqMod.status = StatusProcessing
		signRequestCache.Store(reqId, signReqMod)
	} else {
		return errors.Errorf("you haven't call Sign interface for request id %q, thus no message to sign.", reqId)
	}

	// simulate processing sign, which may take some time, including waiting for enough parties call
	for !m.isEnoughPartiesCalledSign() {
		time.Sleep(3)
	}

	value, _ = signRequestCache.Load(reqId)
	signReqMod := value.(*signRequestModel)
	message := signReqMod.payload.Hash

	value, _ = keyGenRequestsCache.Load(reqId)
	keygenMod := value.(*keyGenRequestModel)
	signer := keygenMod.signer

	sig, err := signer.SignHash([]byte(message))
	if err != nil {
		return errors.WithStack(err)
	}

	signReqMod.signature = string(sig)
	signReqMod.status = StatusDone
	signRequestCache.Store(reqId, signReqMod)

	logger.Debug("sign request DONE",
		logger.Field{"requestId", reqId},
		logger.Field{"signature", signReqMod.signature})
	return nil
}
