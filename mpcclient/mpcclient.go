package mpcclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/avalido/mpc-controller/core/types"
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

type MyMpcClient struct {
	Logger       logger.Logger
	MpcServerUrl string
}

func (c *MyMpcClient) Keygen(ctx context.Context, req *types.KeygenRequest) error {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.MpcServerUrl+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return true, errors.WithStack(err)
		}
		prom.MpcKeygenPosted.Inc()
		return false, nil
	})

	if err != nil {
		c.Logger.Errorf("Failed to send keygen request %+v, error:%+v", req, err)
	} else {
		c.Logger.Debugf("Sent keygen request %+v", req)
	}

	return errors.WithStack(err)
}

func (c *MyMpcClient) Sign(ctx context.Context, req *types.SignRequest) (err error) {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.MpcServerUrl+"/sign", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return true, errors.WithStack(err)
		}
		prom.MpcSignPosted.Inc()
		return false, nil
	})

	if err != nil {
		c.Logger.Errorf("Failed to send sign request %+v, error:%+v", req, err)
	} else {
		c.Logger.Debugf("Sent sign request %+v", req)
	}

	return errors.WithStack(err)
}

func (c *MyMpcClient) Result(ctx context.Context, reqId string) (*types.Result, error) {
	var payload = strings.NewReader("")
	var resp *http.Response
	var err error
	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = http.Post(c.MpcServerUrl+"/result/"+reqId, "application/json", payload)
		if err != nil {
			return true, errors.Wrap(err, "failed to post request")
		}
		prom.MpcResultPosted.Inc()
		return false, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var res types.Result
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse mpc result")
	}
	if res.Status == types.StatusDone {
		prom.MpcResulDone.Inc()
		switch res.Type {
		case types.TypKeygen:
			prom.MpcKeygenDone.Inc()
		case types.TypSign:
			prom.MpcSignDone.Inc()
		}
	}
	return &res, nil
}

type SimulatingMpcClient struct {
	Logger       logger.Logger
	privateKey   *avaCrypto.PrivateKeySECP256K1R
	keygenReqMap map[string]struct{}
	signReqMap   map[string]*types.SignRequest // TODO: Add lock
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
		signReqMap:   map[string]*types.SignRequest{},
	}, nil
}

func (c *SimulatingMpcClient) UncompressedPublicKeyBytes() []byte {
	pk := c.privateKey.PublicKey().Bytes()
	pkEcdsa, _ := ethCrypto.DecompressPubkey(pk)
	pkb := ethCrypto.FromECDSAPub(pkEcdsa)
	return pkb
}

func (c *SimulatingMpcClient) Keygen(ctx context.Context, req *types.KeygenRequest) error {
	c.keygenReqMap[req.ReqID] = struct{}{}
	return nil
}

func (c *SimulatingMpcClient) Sign(ctx context.Context, req *types.SignRequest) error {
	c.signReqMap[req.ReqID] = req
	return nil
}

func (c *SimulatingMpcClient) Result(ctx context.Context, reqID string) (*types.Result, error) {
	if _, ok := c.keygenReqMap[reqID]; ok {
		return &types.Result{
			ReqID: reqID,

			Result: hex.EncodeToString(c.privateKey.PublicKey().Bytes()),
			Type:   types.TypKeygen,
			Status: types.StatusDone,
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
		return &types.Result{
			ReqID:  reqID,
			Result: hex.EncodeToString(sig),
			Type:   types.TypSign,
			Status: types.StatusDone,
		}, nil
	}
	return nil, nil
}
