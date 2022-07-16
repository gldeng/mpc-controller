package staking

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"strconv"
)

// todo: consider reuse /mpc-controller/core/signer.Signer

type SignRequester struct {
	core.SignDoner
	SignRequestArgs
}

type SignRequestArgs struct {
	TaskID                 string
	CompressedPartiPubKeys []string
	CompressedGenPubKeyHex string
}

func (s *SignRequester) SignExportTx(ctx context.Context, exportTxHash []byte) ([65]byte, error) {
	exportTxSignReq := core.SignRequest{
		SignReqID:              s.TaskID + "-" + strconv.Itoa(0),
		CompressedGenPubKeyHex: s.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	res, err := s.SignDone(ctx, &exportTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export tx, RequestID: %q", exportTxSignReq.SignReqID)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *SignRequester) SignImportTx(ctx context.Context, importTxHash []byte) ([65]byte, error) {
	importTxSignReq := core.SignRequest{
		SignReqID:              s.TaskID + "-" + strconv.Itoa(1),
		CompressedGenPubKeyHex: s.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	res, err := s.SignDone(ctx, &importTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export RequestID: %q", importTxSignReq.SignReqID)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *SignRequester) SignAddDelegatorTx(ctx context.Context, addDelegatorTxHash []byte) ([65]byte, error) {
	addDelegatorTxSignReq := core.SignRequest{
		SignReqID:              s.TaskID + "-" + strconv.Itoa(2),
		CompressedGenPubKeyHex: s.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(addDelegatorTxHash),
	}

	res, err := s.SignDone(ctx, &addDelegatorTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export RequestID: %q", addDelegatorTxSignReq.SignReqID)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}
