package stake

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/pkg/errors"
)

const (
	StatusStarted Status = iota
	StatusBuilt
	StatusSent
	StatusOK
)

type Status int

type Task struct {
	Ctx    context.Context
	Logger logger.Logger

	DB   storage.DB
	Pool pool.WorkerPool

	Bound transactor.Transactor

	TriggerReq  *events.StakeRequestAdded
	PartiPubKey storage.PubKey

	status Status

	joinReq *storage.JoinRequest
}

func (t *Task) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *Task) do() bool {
	switch t.status {
	case StatusStarted:
		err := t.buildTask()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to build task")
			return false
		}
		t.status = StatusBuilt
	case StatusBuilt:
		if err := t.DB.SaveModel(t.Ctx, t.joinReq); err != nil {
			t.Logger.ErrorOnError(err, "Failed to save joining request data")
			return false
		}

		_, _, err := t.Bound.JoinRequest(t.Ctx, t.joinReq.PartiId, t.joinReq.ReqHash)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to send JoinRequest")
			return false
		}

		t.status = StatusOK
		t.Logger.Info("Joined stake request", []logger.Field{{"joinedStakeReq",
			fmt.Sprintf("reqNo:%v, reqHash:%v", t.TriggerReq.RequestNumber.Uint64(), t.joinReq.ReqHash.String())}}...)
		return false
	// todo: make it async
	case StatusSent:
		// todo: check async status
	}
	return true
}

func (t *Task) buildTask() error {
	genPubKey := &storage.GeneratedPublicKey{}
	key := genPubKey.KeyFromHash(t.TriggerReq.PublicKey)
	err := t.DB.MGet(t.Ctx, key, genPubKey)
	if err != nil {
		return errors.WithStack(err)
	}

	participant := storage.Participant{
		PubKey:  hash256.FromBytes(t.PartiPubKey),
		GroupId: genPubKey.GroupId,
	}

	err = t.DB.LoadModel(t.Ctx, &participant)
	if err != nil {
		return errors.WithStack(err)
	}

	partiId := participant.ParticipantId()
	txHash := t.TriggerReq.Raw.TxHash

	stakeReq := storage.StakeRequest{
		ReqNo:              t.TriggerReq.RequestNumber.Uint64(),
		TxHash:             txHash,
		NodeID:             t.TriggerReq.NodeID,
		Amount:             t.TriggerReq.Amount.String(),
		StartTime:          t.TriggerReq.StartTime.Int64(),
		EndTime:            t.TriggerReq.EndTime.Int64(),
		GeneratedPublicKey: genPubKey,
	}

	reqHash := stakeReq.ReqHash()
	joinReq := &storage.JoinRequest{
		ReqHash: reqHash,
		PartiId: partiId,
		Args:    &stakeReq,
	}
	t.joinReq = joinReq
	return nil
}
