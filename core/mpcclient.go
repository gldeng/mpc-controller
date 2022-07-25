package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
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

	payloadBytes []byte
	result       *Result
	handled      bool
	err          error
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
	url                string
	log                logger.Logger
	lock               *sync.Mutex
	once               *sync.Once
	pendingSignReqMap  map[string]*SignRequest
	pendingSignReqChan chan *SignRequest
}

func NewMpcClient(log logger.Logger, url string) (*MpcClientImp, error) {
	c := &MpcClientImp{
		url:                url,
		log:                log,
		lock:               new(sync.Mutex),
		once:               new(sync.Once),
		pendingSignReqMap:  make(map[string]*SignRequest),
		pendingSignReqChan: make(chan *SignRequest, 1024),
	}

	return c, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// Request

func (c *MpcClientImp) Keygen(ctx context.Context, request *KeygenRequest) (err error) {
	normalized, err := crypto.NormalizePubKeys(request.CompressedPartiPubKeys)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize public keys")
	}
	request.CompressedPartiPubKeys = normalized
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal KeygenRequest")
	}

	err = backoff.RetryFnExponential100Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.url+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to post keygen request")
	return
}

func (c *MpcClientImp) Sign(ctx context.Context, request *SignRequest) (err error) {
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal SignRequest")
	}
	request.payloadBytes = payloadBytes

	c.lock.Lock()
	c.pendingSignReqMap[request.SignReqID] = request
	c.lock.Unlock()

	c.pendingSignReqChan <- request
	c.log.Debug("Received sign request", []logger.Field{
		{"pendingSignReqMap", len(c.pendingSignReqMap)},
		{"pendingSignReqChan", len(c.pendingSignReqChan)}}...)

	c.once.Do(func() {
		go func() {
			for {
				select {
				case <-ctx.Done():
					// todo: take more effective measures to handle left pending sign requests.
					c.log.WarnOnTrue(len(c.pendingSignReqMap) != 0 || len(c.pendingSignReqChan) != 0,
						"Quited but with pending sign requests not handled",
						[]logger.Field{{"pendingSignReqMap", len(c.pendingSignReqMap)},
							{"pendingSignReqChan", len(c.pendingSignReqChan)}}...)
					return
				case signReq := <-c.pendingSignReqChan:
					err = backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
						_, err = http.Post(c.url+"/sign", "application/json", bytes.NewBuffer(signReq.payloadBytes))
						if err != nil {
							return true, errors.WithStack(err)
						}
						return false, nil
					})
					if err != nil {
						c.lock.Lock()
						signReq.handled = true
						signReq.err = errors.Wrapf(err, "failed to post")
						c.lock.Unlock()
						break
					}

					res, err := c.ResultDone(ctx, signReq.SignReqID)
					if err != nil {
						c.lock.Lock()
						signReq.handled = true
						signReq.err = errors.Wrapf(err, "not done as expected")
						c.lock.Unlock()
						break
					}

					c.lock.Lock()
					signReq.handled = true
					signReq.result = res
					c.lock.Unlock()
				}
			}
		}()
	})
	return
}

func (c *MpcClientImp) Result(ctx context.Context, reqId string) (res *Result, err error) {
	payload := strings.NewReader("")
	var res_ *http.Response
	err = backoff.RetryFnExponential100Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		res_, err = http.Post(c.url+"/result/"+reqId, "application/json", payload)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to post result request")
	if err != nil {
		return
	}

	defer res_.Body.Close()
	body, _ := ioutil.ReadAll(res_.Body)
	res = new(Result)
	err = json.Unmarshal(body, &res)
	if err != nil {
		err = errors.Wrapf(err, "failed to unmarshal response body")
		return
	}
	return
}

// ---------------------------------------------------------------------------------------------------------------------
// Request and wait until it's DONE
// todo: add retry times limit

func (c *MpcClientImp) KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error) {
	err = c.Keygen(ctx, request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	//time.Sleep(time.Second * 5)
	res, err = c.ResultDone(ctx, request.KeygenReqID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return
}

func (c *MpcClientImp) SignDone(ctx context.Context, request *SignRequest) (res *Result, err error) {
	err = c.Sign(ctx, request)
	if err != nil {
		err = errors.Wrapf(err, fmt.Sprintf("failed to sign request %v", request.SignReqID))
		return
	}
	err = backoff.RetryFnExponential100Times(ctx, time.Second, time.Second*10, func() (retry bool, err error) {
		var signReqCopy SignRequest
		c.lock.Lock()
		signReq := c.pendingSignReqMap[request.SignReqID]
		signReqCopy = *signReq
		c.lock.Unlock()
		if !signReqCopy.handled {
			return true, nil
		}
		res = signReqCopy.result
		err = signReqCopy.err
		c.lock.Lock()
		delete(c.pendingSignReqMap, request.SignReqID)
		c.lock.Unlock()
		return false, err
	})

	err = errors.Wrapf(err, fmt.Sprintf("failed to sign request %v", request.SignReqID))
	return
}

func (c *MpcClientImp) ResultDone(ctx context.Context, mpcReqId string) (res *Result, err error) {
	err = backoff.RetryFnExponential100Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		res, err = c.Result(ctx, mpcReqId)
		if err != nil {
			return true, errors.WithStack(err)
		}
		if strings.Contains(res.RequestStatus, "ERROR") {
			return false, errors.Wrap(&ErrTypSignErr{ErrMsg: res.RequestStatus}, "request not done")
		}
		if res.RequestStatus != "DONE" {
			return true, errors.New(res.RequestStatus)
		}
		if res.Result == "" {
			return false, errors.WithStack(&ErrTypEmptySignResult{ErrMsg: "sign is done but got empty result"})
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to request result or request not done for request id %q", mpcReqId)
	return
}
