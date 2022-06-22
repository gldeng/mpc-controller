package staking

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"strconv"
)

// todo: consider refactoring with Template Method design pattern

type SignRequester struct {
	core.SignDoner
	SignRequestArgs
}

type SignRequestArgs struct {
	TaskID                    string
	NormalizedParticipantKeys []string
	PubKeyHex                 string
}

func (s *SignRequester) SignExportTx(ctx context.Context, exportTxHash []byte) ([65]byte, error) {
	exportTxSignReq := core.SignRequest{
		RequestId:       s.TaskID + "-" + strconv.Itoa(0),
		PublicKey:       s.PubKeyHex,
		ParticipantKeys: s.NormalizedParticipantKeys,
		Hash:            bytes.BytesToHex(exportTxHash),
	}

	spew.Dump(exportTxSignReq)

	res, err := s.SignDone(ctx, &exportTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export tx, RequestID: %q", exportTxSignReq.RequestId)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *SignRequester) SignImportTx(ctx context.Context, importTxHash []byte) ([65]byte, error) {
	importTxSignReq := core.SignRequest{
		RequestId:       s.TaskID + "-" + strconv.Itoa(1),
		PublicKey:       s.PubKeyHex,
		ParticipantKeys: s.NormalizedParticipantKeys,
		Hash:            bytes.BytesToHex(importTxHash),
	}
	spew.Dump(importTxSignReq)

	res, err := s.SignDone(ctx, &importTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export RequestID: %q", importTxSignReq.RequestId)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *SignRequester) SignAddDelegatorTx(ctx context.Context, addDelegatorTxHash []byte) ([65]byte, error) {
	addDelegatorTxSignReq := core.SignRequest{
		RequestId:       s.TaskID + "-" + strconv.Itoa(2),
		PublicKey:       s.PubKeyHex,
		ParticipantKeys: s.NormalizedParticipantKeys,
		Hash:            bytes.BytesToHex(addDelegatorTxHash),
	}
	spew.Dump(addDelegatorTxSignReq)

	res, err := s.SignDone(ctx, &addDelegatorTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export RequestID: %q", addDelegatorTxSignReq.RequestId)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}
