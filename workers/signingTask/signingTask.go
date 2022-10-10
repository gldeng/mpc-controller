package signingTask

import (
	"context"
	"github.com/alitto/pond"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
)

const (
	StatusCreated Status = iota
	StatusSubmitted
	StatusOK
	StatusFailed
)

type Status int

type SigningTask struct {
	Status Status
	Ctx    context.Context
	Logger logger.Logger
	Signer core.Signer

	SignReq *core.SignRequest

	*pond.WorkerPool
}

func (t *SigningTask) Do() {
	switch t.Status {
	case StatusCreated:
		//err := t.Signer.Sign(t.Ctx, t.SignReq)
		//if err != nil {
		//	t.Logger.ErrorOnError(err, "")
		//}
		//
		//// Store data for later reuse
		//joinReq := &storage.JoinRequest{
		//	ReqHash: t.ReqHash,
		//	PartiId: t.PartiID,
		//}
		//
		//if t.Type == storage.TaskTypStake && t.StakeReq != nil {
		//	joinReq.Args = t.StakeReq
		//} else if t.Type == storage.TaskTypRecover && t.RecoverReq != nil {
		//	joinReq.Args = t.RecoverReq
		//}
		//
		//if joinReq.Args == nil {
		//	t.Logger.Error("Joining data not provided")
		//	return
		//}
		//
		//if err := t.DB.SaveModel(t.Ctx, joinReq); err != nil {
		//	t.Logger.ErrorOnError(err, "Failed to save joining request data")
		//	return
		//}
		//
		//// Send request
		//_, _, err := t.Transactor.JoinRequest(t.Ctx, t.PartiID, t.ReqHash)
		//if err != nil {
		//	t.Logger.DebugOnError(err, "Joining task failed", []logger.Field{{"joiningTaskFailed",
		//		fmt.Sprintf("type:%v reqHash:%v args:%v", t.Type, t.ReqHash, joinReq.Args)}}...)
		//	return
		//}

		t.Status = StatusOK

		// todo: make it async

		// todo: consider adding metrics
	case StatusSubmitted:
		// todo: check async status
	}
}
