package core

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/backoff"
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

	c.Logger.ErrorOnError(err, "Failed to send KeygenRequest", []logger.Field{{"keygenReq", req}}...)
	c.Logger.DebugNilError(err, "Send KeygenRequest", []logger.Field{{"keygenReq", req}}...)
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

	c.Logger.ErrorOnError(err, "Failed to send SignRequest", []logger.Field{{"signReq", req}}...)
	c.Logger.DebugNilError(err, "Sent SignRequest", []logger.Field{{"signReq", req}}...)
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
