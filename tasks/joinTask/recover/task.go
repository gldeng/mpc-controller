package stake

import (
	"context"
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

	UTXOToRecover *events.UTXOFetched
	PartiPubKey   storage.PubKey

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

		t.Logger.Info("Joined UTXO recover")
		return false
	// todo: make it async
	case StatusSent:
		// todo: check async status
	}
	return true
}

func (t *Task) buildTask() error {
	genPubKey := storage.GeneratedPublicKey{
		GenPubKey: t.UTXOToRecover.GenPubKey,
	}
	err := t.DB.LoadModel(t.Ctx, &genPubKey)
	if err != nil {
		return errors.Wrapf(err, "failed to load generated public key")
	}

	participant := storage.Participant{
		PubKey:  hash256.FromBytes(t.PartiPubKey),
		GroupId: genPubKey.GroupId,
	}

	err = t.DB.LoadModel(t.Ctx, &participant)
	if err != nil {
		return errors.Wrapf(err, "failed to load participant")
	}

	recoverUTXOReq := storage.RecoverRequest{
		TxID:               t.UTXOToRecover.UTXO.TxID,
		OutputIndex:        t.UTXOToRecover.UTXO.OutputIndex,
		GeneratedPublicKey: &genPubKey,
	}

	partiId := participant.ParticipantId()
	reqHash := recoverUTXOReq.ReqHash()

	joinReq := &storage.JoinRequest{
		ReqHash: reqHash,
		PartiId: partiId,
		Args:    &recoverUTXOReq,
	}
	t.joinReq = joinReq
	return nil
}
