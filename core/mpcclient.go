package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/backoff"
	mpcErrors "github.com/avalido/mpc-controller/utils/errors"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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
	controllerID   string
	url            string
	log            logger.Logger
	sentSignReqs   uint32
	unsentSignReqs uint32
	doneSignReqs   uint32
	errorSignReqs  uint32
	once           *sync.Once
	lock           sync.Mutex
}

func NewMpcClient(log logger.Logger, url, controllerID string) (*MpcClientImp, error) {
	c := &MpcClientImp{
		controllerID: controllerID,
		url:          url,
		log:          log,
		once:         new(sync.Once),
	}

	return c, nil
}

func (c *MpcClientImp) KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error) {
	err = c.Keygen(ctx, request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res, err = c.ResultDone(ctx, request.KeygenReqID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return
}

func (c *MpcClientImp) Keygen(ctx context.Context, request *KeygenRequest) (err error) {
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal KeygenRequest")
	}

	err = backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.url+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		c.log.InfoNilError(err, "Posted keygen request", []logger.Field{{"postedKeygenReq", request}}...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		prom.KeygenRequestPosted.Add(1)
		return false, nil
	})
	err = errors.Wrapf(err, "failed to post keygen request")
	return
}

func (c *MpcClientImp) SignDone(ctx context.Context, request *SignRequest) (res *Result, err error) {
	defer func() {
		c.log.Debug("Sign request stats", []logger.Field{
			{"controllerID", c.controllerID},
			{"sentSignReqs", atomic.LoadUint32(&c.sentSignReqs)},
			{"doneSignReqs", atomic.LoadUint32(&c.doneSignReqs)},
			{"unsentSignReqs", atomic.LoadUint32(&c.unsentSignReqs)},
			{"errorSignReqs", atomic.LoadUint32(&c.errorSignReqs)}}...)
	}()

	err = c.Sign(ctx, request)
	if err != nil {
		atomic.AddUint32(&c.unsentSignReqs, 1)
		err = errors.Wrapf(err, fmt.Sprintf("failed to requst sign %v", request.SignReqID))
		return
	}
	atomic.AddUint32(&c.sentSignReqs, 1)

	res, err = c.ResultDone(ctx, request.SignReqID)
	if err != nil {
		c.log.ErrorOnError(err, "Sign request got error", []logger.Field{{"signRes", res}, {"signReq", request}}...)
		atomic.AddUint32(&c.errorSignReqs, 1)
		return res, errors.WithStack(err)
	}
	if res == nil {
		atomic.AddUint32(&c.errorSignReqs, 1)
		return nil, errors.Errorf("Got nil result for sign request %v", request.SignReqID)
	}
	if res.Result == "" {
		atomic.AddUint32(&c.errorSignReqs, 1)
		return res, errors.WithStack(mpcErrors.Errorf(&ErrTypEmptySignResult{}, "got result for sign request %v but it's content is empty", request.SignReqID))
	}

	atomic.AddUint32(&c.doneSignReqs, 1)
	return
}

func (c *MpcClientImp) Sign(ctx context.Context, request *SignRequest) (err error) {
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal SignRequest")
	}

	err = backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() (bool, error) {
		c.lock.Lock()
		defer c.lock.Unlock()
		resp, err := http.Post(c.url+"/sign", "application/json", bytes.NewBuffer(payloadBytes)) // todo: check response?
		c.log.InfoNilError(err, "Posted sign request", []logger.Field{{"postedSignReq", request}}...)

		if resp.StatusCode != 200 {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			c.log.Info("Failed post sign request", []logger.Field{{"error", body}}...)
		}

		if err != nil {
			return true, errors.WithStack(err)
		}
		switch {
		case strings.Contains(request.SignReqID, "STAKE"):
			prom.StakeSignTaskAdded.Inc()
		case strings.Contains(request.SignReqID, "PRINCIPAL"):
			prom.PrincipalUTXOSignTaskAdded.Inc()
		case strings.Contains(request.SignReqID, "REWARD"):
			prom.RewardUTXOSignTaskAdded.Inc()
		}
		return false, nil
	})

	err = errors.Wrapf(err, "failed to request sign")
	return
}

func (c *MpcClientImp) ResultDone(ctx context.Context, mpcReqId string) (res *Result, err error) {
	err = backoff.RetryFnExponential100Times(c.log, ctx, time.Second, time.Second*10, func() (bool, error) {
		res, err = c.Result(ctx, mpcReqId)
		if err != nil {
			return false, errors.WithStack(err)
		}
		if strings.Contains(res.RequestStatus, "ERROR") {
			return false, errors.Wrap(&ErrTypSignErr{ErrMsg: res.RequestStatus}, "error result")
		}
		if res.RequestStatus == "DONE" && res.Result != "" {
			switch {
			case strings.Contains(mpcReqId, "STAKE"):
				prom.StakeSignTaskDone.Inc()
			case strings.Contains(mpcReqId, "PRINCIPAL"):
				prom.PrincipalUTXOSignTaskDone.Inc()
			case strings.Contains(mpcReqId, "REWARD"):
				prom.RewardUTXOSignTaskDone.Inc()
			}
			return false, nil
		}
		return true, errors.New(res.RequestStatus)
	})
	err = errors.Wrapf(err, "failed to request result or got error result for request controllerID %q", mpcReqId)
	return
}

func (c *MpcClientImp) Result(ctx context.Context, reqId string) (res *Result, err error) {
	payload := strings.NewReader("")
	var resp *http.Response
	err = backoff.RetryFnExponential10Times(c.log, ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = http.Post(c.url+"/result/"+reqId, "application/json", payload)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to request result")
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	res = new(Result)
	err = json.Unmarshal(body, &res)
	err = errors.Wrapf(err, "failed to unmarshal response body")
	return
}
