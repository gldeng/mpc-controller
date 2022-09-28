package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/dispatcher"
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
	ReqID                  string   `json:"request_id"`
	CompressedPartiPubKeys []string `json:"public_keys"`
	Threshold              uint64   `json:"t"`
}

type SignRequest struct {
	ReqID                  string          `json:"request_id"`
	Kind                   events.SignKInd `json:"-"`
	Hash                   string          `json:"message"`
	CompressedGenPubKeyHex string          `json:"public_key"`
	CompressedPartiPubKeys []string        `json:"participant_public_keys"`
}

type Result struct {
	ReqID     string           `json:"request_id"`
	Result    string           `json:"result"`
	ReqType   events.ReqType   `json:"request_type"`
	ReqStatus events.ReqStatus `json:"request_status"`
}

var _ MpcClient = (*MpcClientImp)(nil)

type MpcClient interface {
	Keygen(ctx context.Context, keygenReq *KeygenRequest) error
	Sign(ctx context.Context, signReq *SignRequest) error
	Result(ctx context.Context, reqID string) (*Result, error)
}

type MpcClientImp struct {
	Logger       logger.Logger
	ControllerID string
	MpcServerUrl string
	Publisher    dispatcher.Publisher

	sentSignReqs   uint32
	unsentSignReqs uint32
	doneSignReqs   uint32
	errorSignReqs  uint32
	once           *sync.Once
	lock           sync.Mutex
	pendingReqs    *sync.Map // todo: persistence and recovery
}

func (c *MpcClientImp) Init(ctx context.Context) {
	go c.checkResult(ctx)
}

func (c *MpcClientImp) KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error) {
	err = c.Keygen(ctx, request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res, err = c.ResultDone(ctx, request.ReqID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return
}

func (c *MpcClientImp) Keygen(ctx context.Context, req *KeygenRequest) (err error) {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal KeygenRequest")
	}

	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		_, err = http.Post(c.MpcServerUrl+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
		c.Logger.InfoNilError(err, "Posted keygen request", []logger.Field{{"postedKeygenReq", req}}...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		prom.KeygenRequestPosted.Inc()
		c.pendingReqs.Store(req.ReqID, req)
		return false, nil
	})
	err = errors.Wrapf(err, "failed to post keygen request")
	return
}

func (c *MpcClientImp) SignDone(ctx context.Context, request *SignRequest) (res *Result, err error) {
	defer func() {
		c.Logger.Debug("Sign request stats", []logger.Field{
			{"controllerID", c.ControllerID},
			{"sentSignReqs", atomic.LoadUint32(&c.sentSignReqs)},
			{"doneSignReqs", atomic.LoadUint32(&c.doneSignReqs)},
			{"unsentSignReqs", atomic.LoadUint32(&c.unsentSignReqs)},
			{"errorSignReqs", atomic.LoadUint32(&c.errorSignReqs)}}...)
	}()

	err = c.Sign(ctx, request)
	if err != nil {
		atomic.AddUint32(&c.unsentSignReqs, 1)
		err = errors.Wrapf(err, fmt.Sprintf("failed to requst sign %v", request.ReqID))
		return
	}
	atomic.AddUint32(&c.sentSignReqs, 1)

	res, err = c.ResultDone(ctx, request.ReqID)
	if err != nil {
		c.Logger.ErrorOnError(err, "Sign request got error", []logger.Field{{"signRes", res}, {"signReq", request}}...)
		atomic.AddUint32(&c.errorSignReqs, 1)
		return res, errors.WithStack(err)
	}
	if res == nil {
		atomic.AddUint32(&c.errorSignReqs, 1)
		return nil, errors.Errorf("Got nil result for sign request %v", request.ReqID)
	}
	if res.Result == "" {
		atomic.AddUint32(&c.errorSignReqs, 1)
		return res, errors.WithStack(mpcErrors.Errorf(&ErrTypEmptySignResult{}, "got result for sign request %v but it's content is empty", request.ReqID))
	}

	atomic.AddUint32(&c.doneSignReqs, 1)
	return
}

func (c *MpcClientImp) Sign(ctx context.Context, req *SignRequest) (err error) {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal SignRequest")
	}

	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		c.lock.Lock()
		defer c.lock.Unlock()
		_, err = http.Post(c.MpcServerUrl+"/sign", "application/json", bytes.NewBuffer(payloadBytes)) // todo: check response?
		c.Logger.InfoNilError(err, "Posted sign request", []logger.Field{{"postedSignReq", req}}...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		switch {
		case strings.Contains(req.ReqID, "STAKE"):
			prom.StakeSignTaskAdded.Inc()
		case strings.Contains(req.ReqID, "PRINCIPAL"):
			prom.PrincipalUTXOSignTaskAdded.Inc()
		case strings.Contains(req.ReqID, "REWARD"):
			prom.RewardUTXOSignTaskAdded.Inc()
		}

		c.pendingReqs.Store(req.ReqID, req)
		return false, nil
	})

	err = errors.Wrapf(err, "failed to request sign")
	return
}

func (c *MpcClientImp) ResultDone(ctx context.Context, mpcReqId string) (res *Result, err error) {
	err = backoff.RetryFnExponential100Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		res, err = c.Result(ctx, mpcReqId)
		if err != nil {
			return false, errors.WithStack(err)
		}
		if strings.Contains(string(res.ReqStatus), "ERROR") {
			return false, errors.Wrap(&ErrTypSignErr{ErrMsg: string(res.ReqStatus)}, "error result")
		}
		if res.ReqStatus == "DONE" && res.Result != "" {
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
		return true, errors.New(string(res.ReqStatus))
	})
	err = errors.Wrapf(err, "failed to request result or got error result for request controllerID %q", mpcReqId)
	return
}

func (c *MpcClientImp) checkResult(ctx context.Context) {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.pendingReqs.Range(func(key, value any) bool {
				// Fetch result
				reqId := key.(string)
				resp, err := http.Post(c.MpcServerUrl+"/result/"+reqId, "application/json", strings.NewReader(""))
				if err != nil {
					c.Logger.ErrorOnError(errors.WithStack(err), "failed to post request")
					return false
				}
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				res := new(Result)
				_ = json.Unmarshal(body, &res)

				if res.ReqStatus == events.ReqStatusDone && res.Result != "" {
					// Publish result
					var evt interface{}
					switch {
					case res.ReqType == events.ReqTypSignSign:
						myEvt := events.SignDone{
							ReqID:  res.ReqID,
							Result: new(events.Signature).FromHex(res.Result),
						}
						evt = &myEvt
					}
					c.Publisher.Publish(ctx, dispatcher.NewEvtObj(evt, nil))

					// Record metrics
					switch {
					case strings.Contains(reqId, string(events.ReqIDPrefixKeygen)):
						prom.KeygenRequestDone.Inc()
					case strings.Contains(reqId, string(events.ReqIDPrefixSignStake)):
						prom.StakeSignTaskDone.Inc()
					case strings.Contains(reqId, string(events.ReqIDPrefixSignPrincipal)):
						prom.PrincipalUTXOSignTaskDone.Inc()
					case strings.Contains(reqId, string(events.ReqIDPrefixSignReward)):
						prom.RewardUTXOSignTaskDone.Inc()
					}

					// Delete pending requests
					c.pendingReqs.Delete(reqId)

					c.Logger.Info("Fetched result from mpc-server", []logger.Field{{"mpcServerResultDone", res}}...)
				} else {
					c.Logger.Debug("Fetched result from mpc-server", []logger.Field{{"mpcServerResultFetched", res}}...)
				}
				return true
			})
		}
	}
}

func (c *MpcClientImp) Result(ctx context.Context, reqId string) (res *Result, err error) {
	payload := strings.NewReader("")
	var resp *http.Response
	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = http.Post(c.MpcServerUrl+"/result/"+reqId, "application/json", payload)
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
