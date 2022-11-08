package core

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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
	CompressedGenPubKeyHex string   `json:"public_key"`
	CompressedPartiPubKeys []string `json:"participant_public_keys"`
}

type Result struct {
	ReqID  string `json:"request_id"`
	Result string `json:"result"`
	Type   Type   `json:"request_type"`
	Status Status `json:"request_status"`
}

// todo: Prometheus metrics

type MpcClient interface {
	Keygen(ctx context.Context, req *KeygenRequest) error
	Sign(ctx context.Context, req *SignRequest) error
	Result(ctx context.Context, reqID string) (*Result, error)
}

type MyMpcClient struct {
	Logger       logger.Logger
	MpcServerUrl string
}

func (c *MyMpcClient) Keygen(ctx context.Context, req *KeygenRequest) error {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.MpcServerUrl+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return true, errors.WithStack(err)
		}
		prom.KeygenRequestPosted.Inc()
		return false, nil
	})

	if err != nil {
		c.Logger.Errorf("Failed to send keygen request %+v, error:%+v", req, err)
	} else {
		c.Logger.Debugf("Sent keygen request %+v", req)
	}

	return errors.WithStack(err)
}

func (c *MyMpcClient) Sign(ctx context.Context, req *SignRequest) (err error) {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.MpcServerUrl+"/sign", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return true, errors.WithStack(err)
		}

		return false, nil
	})

	if err != nil {
		c.Logger.Errorf("Failed to send sign request %+v, error:%+v", req, err)
	} else {
		c.Logger.Debugf("Sent sign request %+v", req)
	}

	return errors.WithStack(err)
}

func (c *MyMpcClient) Result(ctx context.Context, reqId string) (*Result, error) {
	var payload = strings.NewReader("")
	var resp *http.Response
	var err error
	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = http.Post(c.MpcServerUrl+"/result/"+reqId, "application/json", payload)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var res Result
	err = json.Unmarshal(body, &res)
	return &res, errors.WithStack(err)
}

type SimulatingMpcClient struct {
	Logger       logger.Logger
	privateKey   *avaCrypto.PrivateKeySECP256K1R
	keygenReqMap map[string]struct{}
	signReqMap   map[string]*SignRequest // TODO: Add lock
}

func NewSimulatingMpcClient(privateKey string) (*SimulatingMpcClient, error) {
	sk, err := crypto.ParsePrivateKeySECP256K1R(privateKey)
	if err != nil {
		return nil, err
	}
	return &SimulatingMpcClient{
		Logger:       nil,
		privateKey:   sk,
		keygenReqMap: map[string]struct{}{},
		signReqMap:   map[string]*SignRequest{},
	}, nil
}

func (c *SimulatingMpcClient) UncompressedPublicKeyBytes() []byte {
	pk := c.privateKey.PublicKey().Bytes()
	pkEcdsa, _ := ethCrypto.DecompressPubkey(pk)
	pkb := ethCrypto.FromECDSAPub(pkEcdsa)
	return pkb
}

func (c *SimulatingMpcClient) Keygen(ctx context.Context, req *KeygenRequest) error {
	c.keygenReqMap[req.ReqID] = struct{}{}
	return nil
}

func (c *SimulatingMpcClient) Sign(ctx context.Context, req *SignRequest) error {
	c.signReqMap[req.ReqID] = req
	return nil
}

func (c *SimulatingMpcClient) Result(ctx context.Context, reqID string) (*Result, error) {
	if _, ok := c.keygenReqMap[reqID]; ok {
		return &Result{
			ReqID: reqID,

			Result: hex.EncodeToString(c.privateKey.PublicKey().Bytes()),
			Type:   TypKeygen,
			Status: StatusDone,
		}, nil
	} else if s, ok := c.signReqMap[reqID]; ok {
		hashBytes, err := hex.DecodeString(s.Hash)
		if err != nil {
			return nil, err
		}
		sig, err := c.privateKey.SignHash(hashBytes)
		if err != nil {
			return nil, err
		}
		return &Result{
			ReqID:  reqID,
			Result: hex.EncodeToString(sig),
			Type:   TypSignSign,
			Status: StatusDone,
		}, nil
	}
	return nil, nil
}
