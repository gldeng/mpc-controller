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
)

type KeygenRequest struct {
	RequestId       string   `json:"request_id"`
	ParticipantKeys []string `json:"public_keys"`
	Threshold       uint64   `json:"t"`
}

type SignRequest struct {
	RequestId       string   `json:"request_id"`
	PublicKey       string   `json:"public_key"`
	ParticipantKeys []string `json:"participant_public_keys"`
	Hash            string   `json:"message"`
}

type Result struct {
	RequestId     string `json:"request_id"`
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
	normalized, err := crypto.NormalizePubKeys(request.ParticipantKeys)
	if err != nil {
		c.log.Error("Failed to normalize public keys", logger.Field{"error", err})
		return errors.WithStack(err)
	}
	request.ParticipantKeys = normalized
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		c.log.Error("Failed to marshal KeygenRequest", logger.Field{"error", err})
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponentialForever(c.log, ctx, func() error {
		_, err = http.Post(c.url+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			c.log.Error("Failed to post keygen request", logger.Field{"error", err})
			return errors.WithStack(err)
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
		c.log.Error("Failed to marshal SignRequest", logger.Field{"error", err})
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponentialForever(c.log, ctx, func() error {
		_, err = http.Post(c.url+"/sign", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			c.log.Error("Failed to post sign request", logger.Field{"error", err})
			return errors.WithStack(err)
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

	err := backoff.RetryFnExponentialForever(c.log, ctx, func() error {
		res_, err := http.Post(c.url+"/result/"+reqId, "application/json", payload)
		if err != nil {
			c.log.Error("Failed to post result request", logger.Field{"error", err})
			return errors.WithStack(err)
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
		c.log.Error("Failed to unmarshal response body", logger.Field{"error", err})
		return nil, errors.WithStack(err)

	}
	return &result, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// Request and wait until it's DONE

func (c *MpcClientImp) KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error) {
	err = c.Keygen(ctx, request)
	if err != nil {
		return
	}

	res, err = c.ResultDone(ctx, request.RequestId)
	return
}

func (c *MpcClientImp) SignDone(ctx context.Context, request *SignRequest) (res *Result, err error) {
	err = c.Sign(ctx, request)
	if err != nil {
		return
	}

	res, err = c.ResultDone(ctx, request.RequestId)
	return
}

func (c *MpcClientImp) ResultDone(ctx context.Context, reqId string) (res *Result, err error) {
	err = backoff.RetryFnExponentialForever(c.log, ctx, func() error {
		res, err = c.Result(ctx, reqId)
		if err != nil {
			return errors.WithStack(err)
		}

		if res.RequestStatus != "DONE" {
			c.log.Warn("Request not Done.", []logger.Field{{"reqId", reqId}}...)
			return errors.Errorf("Request not DONE. ReqId: %q", reqId)
		}
		return nil
	})

	return
}
