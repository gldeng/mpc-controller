package mpc_client

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/rand"
	"strings"
	"sync"
	"time"
)

//var globalSigner crypto.Signer

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

// todo: take measures to avoid gorutine leak

func New(parties, threshold int) core.MPCClient {
	m := &MpcClientMock{parties, threshold}
	go func() {
		for {
			select {
			case reqId := <-keygenTaskQueue:
				go func(reqId string) {
					err := m.keygen(reqId)
					if err != nil {
						logger.Error("Failed to generate key",
							logger.Field{"requestId", reqId},
							logger.Field{"error", err})
					}
				}(reqId)

			case reqId := <-signTaskQueue:
				go func(reqId string) {
					err := m.sign(reqId)
					if err != nil {
						logger.Error("Failed to sign message",
							logger.Field{"requestId", reqId},
							logger.Field{"error", err})
					}
				}(reqId)
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
	case lastReqNum == m.parties-1:
		IncreaseRequestNumber(reqId, TypeKeygen)
		keygenTaskQueue <- reqId
	default:
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

	ids := strings.Split(reqId, "-")
	logger.Debug("Mpc-server received sign request", logger.Field{"signReq", signReq})
	for {
		value, ok := keyGenRequestsCache.Load(ids[0])
		if !ok {
			logger.Debug("Mpc-server waiting for corresponding key generated")
			time.Sleep(3)
			continue
		}

		keyGenReqMod := value.(*keyGenRequestModel)
		if keyGenReqMod.status != StatusDone || keyGenReqMod.signer == nil {
			logger.Debug("Mpc-server waiting for corresponding key generated")
			time.Sleep(3)
			continue
		}
		break
	}

	lastReqNum := QueryRequestNumber(reqId, TypeSign)
	switch {
	case lastReqNum == 0:
		signRequestCache.Store(reqId, &signRequestModel{
			status:  StatusReceived,
			payload: signReq})
		IncreaseRequestNumber(reqId, TypeSign)
	case lastReqNum == m.threshold:
		IncreaseRequestNumber(reqId, TypeSign)
		signTaskQueue <- reqId
	default:
		IncreaseRequestNumber(reqId, TypeSign)
	}

	logger.Debug("Mpc-Server received sign request..",
		logger.Field{"requestId", reqId},
		logger.Field{"requestNum", lastReqNum + 1},
		logger.Field{"signReq", signReq})
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

	return nil, errors.Errorf("Mpc-server not found result concerning request id %q", reqID)
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

	privKeyHex := common.Bytes2Hex(signer.PrivateKey().Bytes())
	pubKeyHex := common.Bytes2Hex(signer.PublicKey().Bytes())
	reqMod.pubKeyHex = pubKeyHex
	reqMod.status = StatusDone
	keyGenRequestsCache.Store(reqId, reqMod)
	logger.Debug("Mpc-server generated key",
		logger.Field{"reqId", reqId},
		logger.Field{"privKeyHex", privKeyHex},
		logger.Field{"pubKeyHex", pubKeyHex})
	return nil
}

func (m *MpcClientMock) sign(reqId string) error {
	// simulate time elapse before process sign
	logger.Debug("Mpc-server sign request PROCESSING.", logger.Field{"requestId", reqId})
	value, _ := signRequestCache.Load(reqId)
	signReqMod := value.(*signRequestModel)
	signReqMod.status = StatusProcessing
	signRequestCache.Store(reqId, signReqMod)
	rand.Seed(time.Now().UnixNano())
	randNun := rand.Intn(m.parties)
	time.Sleep(time.Second * time.Duration(randNun))

	ids := strings.Split(reqId, "-")
	message := signReqMod.payload.Hash
	value, _ = keyGenRequestsCache.Load(ids[0])
	keygenMod := value.(*keyGenRequestModel)
	signer := keygenMod.signer

	messageBytes := common.Hex2Bytes(message)
	sigBytes, err := signer.SignHash(messageBytes)
	if err != nil {
		return errors.WithStack(err)
	}
	sigHex := common.Bytes2Hex(sigBytes)

	signReqMod.signature = sigHex
	signReqMod.status = StatusDone
	signRequestCache.Store(reqId, signReqMod)

	logger.Debug("Mpc-server signed message",
		logger.Field{"requestId", reqId},
		logger.Field{"messageHex", message},
		logger.Field{"signatureHex", signReqMod.signature})
	return nil
}
