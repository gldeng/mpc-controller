package staking

import (
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"strconv"
)

type stakeTask interface {
	ExportTxHash() ([]byte, error)
	ImportTxHash() ([]byte, error)
	AddDelegatorTxHash() ([]byte, error)
}

type signRequestCreator struct {
	Task   stakeTask
	TaskID string

	NormalizedParticipantKeys []string
	PubKeyHex                 string

	reqNum    uint8
	txHashHex string
}

// Todo: Consider applying State design pattern

func (s *signRequestCreator) createSignRequest() (*core.SignRequest, error) {
	switch s.reqNum {
	case 0:
		txHashBytes, err := s.Task.ExportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ExportTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		s.reqNum++
	case 1:
		txHashBytes, err := s.Task.ImportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ImportTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		s.reqNum++
	case 2:
		txHashBytes, err := s.Task.AddDelegatorTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create AddDelegatorTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)
	}

	request := core.SignRequest{
		RequestId:       s.TaskID + "-" + strconv.Itoa(int(s.reqNum)),
		PublicKey:       s.PubKeyHex,
		ParticipantKeys: s.NormalizedParticipantKeys,
		Hash:            s.txHashHex,
	}

	return &request, nil
}
