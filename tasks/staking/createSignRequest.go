package staking

import (
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"strconv"
)

type StakeTasker interface {
	ExportTxHash() ([]byte, error)
	ImportTxHash() ([]byte, error)
	AddDelegatorTxHash() ([]byte, error)
}

type SignRequestCreator struct {
	TaskID string

	NormalizedParticipantKeys []string
	PubKeyHex                 string

	reqNum    uint8
	txHashHex string
}

// Todo: Consider applying State design pattern

func (s *SignRequestCreator) CreateSignRequest(task StakeTasker) (*core.SignRequest, error) {
	switch s.reqNum {
	case 0:
		txHashBytes, err := task.ExportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ExportTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		s.reqNum++
	case 1:
		txHashBytes, err := task.ImportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ImportTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		s.reqNum++
	case 2:
		txHashBytes, err := task.AddDelegatorTxHash()
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
