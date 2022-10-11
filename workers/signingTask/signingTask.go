package signingTask

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/alitto/pond"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	StatusCreated Status = iota
	StatusSubmitted
)

type Status int

type SigningTask struct {
	Status Status
	Ctx    context.Context
	Logger logger.Logger

	SignURL string
	SignReq *core.SignRequest

	WorkPool   *pond.WorkerPool
	Dispatcher kbcevents.Dispatcher[*core.Result]
}

func (t *SigningTask) Do() {
	switch t.Status {
	case StatusCreated:
		// Submit signing task for mpc-server to process
		payloadBytes, _ := json.Marshal(t.SignReq)
		err := backoff.RetryFnExponential10Times(t.Logger, t.Ctx, time.Second, time.Second*10, func() (bool, error) {
			_, err := http.Post(t.SignURL+"/sign", "application/json", bytes.NewBuffer(payloadBytes))
			if err != nil {
				return true, errors.WithStack(err)
			}
			return false, nil
		})

		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to submit signing request")
			t.WorkPool.Submit(t.Do)
			return
		}

		t.Status = StatusSubmitted
		t.WorkPool.Submit(t.Do)
	case StatusSubmitted:
		// Check signing result
		var resp *http.Response
		err := backoff.RetryFnExponential10Times(t.Logger, t.Ctx, time.Second, time.Second*10, func() (bool, error) {
			var err error
			resp, err = http.Post(t.SignURL+"/result/"+t.SignReq.ReqID, "application/json", strings.NewReader(""))
			if err != nil {
				return true, errors.WithStack(err)
			}
			return false, nil
		})

		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to check signing result")
			t.WorkPool.Submit(t.Do)
			return
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		res := new(core.Result)
		_ = json.Unmarshal(body, &res)

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			t.WorkPool.Submit(t.Do)
		}

		// Emit signing result
		t.Dispatcher.Dispatch(res)
	}
}
