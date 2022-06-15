package staking

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/pkg/errors"
)

type StakeTaskCreatorer interface {
	CreateStakeTask() (*StakeTask, error)
}

type SignRequestCreatorer interface {
	CreateSignRequest(task StakeTasker) (*core.SignRequest, error)
}

type StakeTaskSignRequester struct {
	StakeTaskCreatorer
	SignRequestCreatorer
	core.SignDoner
	crypto.VerifyHasher

	task *StakeTask
}

func (s *StakeTaskSignRequester) Sign(ctx context.Context) (*StakeTask, error) {
	task, err := s.CreateStakeTask()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s.task = task

	for i := 0; i < 3; i++ {
		signReq, err := s.CreateSignRequest(task)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		res, err := s.SignDone(ctx, signReq)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		sigBytes := bytes.HexToBytes(res.Result)
		err = s.setSig(i, sigBytes)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return task, nil
}

func (s *StakeTaskSignRequester) setSig(reqNum int, sigBytes []byte) error {
	switch reqNum {
	case 0:
		hashBytes, _ := s.task.ExportTxHash()
		ok := s.VerifyHash(hashBytes, sigBytes)
		if !ok {
			return errors.New("Failed to verify ExportTxHash signed signature.")
		}

		err := s.task.SetExportTxSig(bytes.BytesTo65Bytes(sigBytes))
		if err != nil {
			return errors.WithStack(err)
		}
	case 1:
		hashBytes, _ := s.task.ImportTxHash()
		ok := s.VerifyHash(hashBytes, sigBytes)
		if !ok {
			return errors.New("Failed to verify ImportTxHash signed signature.")
		}

		err := s.task.SetImportTxSig(bytes.BytesTo65Bytes(sigBytes))
		if err != nil {
			return errors.WithStack(err)
		}
	case 2:
		hashBytes, _ := s.task.AddDelegatorTxHash()
		ok := s.VerifyHash(hashBytes, sigBytes)
		if !ok {
			return errors.New("Failed to verify AddDelegatorTxHash signed signature.")
		}

		err := s.task.SetAddDelegatorTxSig(bytes.BytesTo65Bytes(sigBytes))
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
