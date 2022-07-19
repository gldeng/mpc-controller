package core

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type KeygenRequest struct {
	KeygenReqID            string   `json:"request_id"`
	CompressedPartiPubKeys []string `json:"public_keys"`
	Threshold              uint64   `json:"t"`
}

type SignRequest struct {
	SignReqID              string   `json:"request_id"`
	CompressedGenPubKeyHex string   `json:"public_key"`
	CompressedPartiPubKeys []string `json:"participant_public_keys"`
	Hash                   string   `json:"message"`
}

type Result struct {
	MPCReqID      string `json:"request_id"`
	Result        string `json:"result"`
	RequestType   string `json:"request_type"`
	RequestStatus string `json:"request_status"`
}

var _ MpcClient = (*MpcClientImp)(nil)

type MpcClient interface {
	Keygen(ctx context.Context, keygenReq *KeygenRequest) error
	Sign(ctx context.Context, signReq *SignRequest) error
	Result(ctx context.Context, reqID string) (*Result, error)
}

type MpcClientImp struct {
	url string
	log logger.Logger
}

func NewMpcClient(log logger.Logger, url string) (*MpcClientImp, error) {
	return &MpcClientImp{url, log}, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// Request

func (c *MpcClientImp) Keygen(ctx context.Context, request *KeygenRequest) error {
	normalized, err := crypto.NormalizePubKeys(request.CompressedPartiPubKeys)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize public keys")
	}
	request.CompressedPartiPubKeys = normalized
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal KeygenRequest")
	}

	err = backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() error {
		_, err = http.Post(c.url+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return errors.Wrapf(err, "failed to post keygen request")
		}
		return nil
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *MpcClientImp) Sign(ctx context.Context, request *SignRequest) error {
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal SignRequest")
	}

	err = backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() error {
		_, err = http.Post(c.url+"/sign", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return errors.Wrapf(err, "failed to post sign request")
		}
		return nil
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *MpcClientImp) Result(ctx context.Context, reqId string) (*Result, error) {
	payload := strings.NewReader("")

	var res *http.Response

	err := backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() error {
		res_, err := http.Post(c.url+"/result/"+reqId, "application/json", payload)
		if err != nil {
			return errors.Wrapf(err, "failed to post result request")
		}
		res = res_
		return nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close()

	var result Result
	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal response body")
	}
	return &result, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// Request and wait until it's DONE
// todo: add retry times limit

func (c *MpcClientImp) KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error) {
	err = c.Keygen(ctx, request)
	if err != nil {
		return
	}

	time.Sleep(time.Second * 5)
	res, err = c.ResultDone(ctx, request.KeygenReqID)
	return
}

func (c *MpcClientImp) SignDone(ctx context.Context, request *SignRequest) (res *Result, err error) {
	err = c.Sign(ctx, request)
	if err != nil {
		return
	}

	time.Sleep(time.Second * 5)
	res, err = c.ResultDone(ctx, request.SignReqID)
	return
}

func (c *MpcClientImp) ResultDone(ctx context.Context, mpcReqId string) (res *Result, err error) {
	err = backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() error {
		res, err = c.Result(ctx, mpcReqId)
		if err != nil {
			return errors.Wrapf(err, "failed to check result")
		}

		if res.RequestStatus != "DONE" {
			return errors.Errorf("MPC request not DONE. mpcReqId: %q", mpcReqId)
		}
		return nil
	})

	return
}
