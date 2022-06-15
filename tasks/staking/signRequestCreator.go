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

func (m *signRequestCreator) CreateSignRequest() (*core.SignRequest, error) {
	switch m.reqNum {
	case 0:
		txHashBytes, err := m.Task.ExportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ExportTxHash")
		}
		m.txHashHex = bytes.BytesToHex(txHashBytes)

		m.reqNum++
	case 1:
		txHashBytes, err := m.Task.ImportTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create ImportTxHash")
		}
		m.txHashHex = bytes.BytesToHex(txHashBytes)

		m.reqNum++
	case 2:
		txHashBytes, err := m.Task.AddDelegatorTxHash()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create AddDelegatorTxHash")
		}
		m.txHashHex = bytes.BytesToHex(txHashBytes)
	}

	request := core.SignRequest{
		RequestId:       m.TaskID + "-" + strconv.Itoa(int(m.reqNum)),
		PublicKey:       m.PubKeyHex,
		ParticipantKeys: m.NormalizedParticipantKeys,
		Hash:            m.txHashHex,
	}

	return &request, nil
}
