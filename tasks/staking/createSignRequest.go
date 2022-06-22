package staking

import (
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"strconv"
)

type SignRequestCreator struct {
	TaskID string

	NormalizedParticipantKeys []string
	PubKeyHex                 string

	reqNum    uint8
	txHashHex string
}

// Todo: Consider applying State design pattern

func (s *SignRequestCreator) CreateSignRequest(task TxHashGenerator) (*core.SignRequest, error) {
	var currentReqNum int
	switch s.reqNum {
	case 0:
		txHashBytes, err := task.ExportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ExportTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		currentReqNum = 0
		s.reqNum++
	case 1:
		txHashBytes, err := task.ImportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ImportTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		currentReqNum = 1
		s.reqNum++
	case 2:
		txHashBytes, err := task.AddDelegatorTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create AddDelegatorTxHash")
		}
		s.txHashHex = bytes.BytesToHex(txHashBytes)

		currentReqNum = 2
	}

	request := core.SignRequest{
		RequestId:       s.TaskID + "-" + strconv.Itoa(currentReqNum),
		PublicKey:       s.PubKeyHex,
		ParticipantKeys: s.NormalizedParticipantKeys,
		Hash:            s.txHashHex,
	}

	return &request, nil
}
