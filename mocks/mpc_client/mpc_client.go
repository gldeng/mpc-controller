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

var keygenTaskQueue = make(chan string, 10)
var signTaskQueue = make(chan string, 10)

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
			case reqId := <-keygenTaskQueue:
				err := m.keygen(reqId)
				if err != nil {
					logger.Error("Failed to generate key",
						logger.Field{"requestId", reqId},
						logger.Field{"error", err})
				}
			case reqId := <-signTaskQueue:
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
	// todo: validate input, currently taking assumption that the content will be identical for the same request id
	reqId := keygenReq.RequestId
	lastReqNum := QueryRequestNumber(reqId, TypeKeygen)

	switch {
	case lastReqNum == 0:
		keyGenRequestsCache.Store(reqId, &keyGenRequestModel{
			status:  StatusReceived,
			payload: keygenReq,
		})
		IncreaseRequestNumber(reqId, TypeKeygen)
	case lastReqNum == 2:
		IncreaseRequestNumber(reqId, TypeKeygen)
		keygenTaskQueue <- reqId
	case lastReqNum == 1 || lastReqNum >= 3:
		IncreaseRequestNumber(reqId, TypeKeygen)
	}

	logger.Debug("Keygen request RECEIVED.",
		logger.Field{"requestId", reqId},
		logger.Field{"requestNum", lastReqNum + 1})
	return nil
}

func (m *MpcClientMock) Sign(ctx context.Context, signReq *core.SignRequest) error {
	// todo: validate input, currently taking assumption that the content will be identical for the same request id
	reqId := signReq.RequestId

	// make sure sign request can be triggered only after there is corresponding valid signer.
	value, ok := keyGenRequestsCache.Load(reqId)
	if !ok {
		return errors.Errorf("you haven't call Keygen interface for request id %q, thus there's no signer to sign", reqId)
	}

	keyGenReqMod := value.(*keyGenRequestModel)
	if keyGenReqMod.status != StatusDone || keyGenReqMod.signer == nil {
		return errors.Errorf("request id %q is in keygen process and you should wait for a signer to be available.", reqId)
	}

	lastReqNum := QueryRequestNumber(reqId, TypeSign)
	switch {
	case lastReqNum == 0:
		signRequestCache.Store(reqId, &signRequestModel{
			status:  StatusReceived,
			payload: signReq})
		IncreaseRequestNumber(reqId, TypeSign)
	case lastReqNum == 2:
		IncreaseRequestNumber(reqId, TypeSign)
		signTaskQueue <- reqId
	case lastReqNum == 1 || lastReqNum >= 3:
		IncreaseRequestNumber(reqId, TypeSign)
	}

	logger.Debug("Sign request RECEIVED.",
		logger.Field{"requestId", reqId},
		logger.Field{"requestNum", lastReqNum + 1})
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

func (m *MpcClientMock) keygen(reqId string) error {
	// simulate time elapse before process keygen
	logger.Debug("Keygen request PROCESSING.", logger.Field{"requestId", reqId})
	value, _ := keyGenRequestsCache.Load(reqId)
	reqMod := value.(*keyGenRequestModel)
	reqMod.status = StatusProcessing
	keyGenRequestsCache.Store(reqId, reqMod)
	rand.Seed(time.Now().UnixNano())
	randNun := rand.Intn(m.parties)
	time.Sleep(time.Second * time.Duration(randNun))

	signer, err := crypto.NewSECP256K1RSigner()
	if err != nil {
		return errors.WithStack(err)
	}
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
	logger.Debug("Sign request PROCESSING.", logger.Field{"requestId", reqId})
	value, _ := signRequestCache.Load(reqId)
	signReqMod := value.(*signRequestModel)
	signReqMod.status = StatusProcessing
	signRequestCache.Store(reqId, signReqMod)
	rand.Seed(time.Now().UnixNano())
	randNun := rand.Intn(m.parties)
	time.Sleep(time.Second * time.Duration(randNun))

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
