package signer

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/pkg/errors"
	"strconv"
)

var _ porter.TxSigner = (*Signer)(nil)

type Signer struct {
	core.SignDoner
	SignRequestArgs
}

type SignRequestArgs struct {
	TaskID                 string
	CompressedPartiPubKeys []string
	CompressedGenPubKeyHex string
}

func (s *Signer) SignExportTx(ctx context.Context, exportTxHash []byte) ([65]byte, error) {
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

	if res == nil || res.Result == "" {
		return [65]byte{}, errors.Errorf("exportTx got no signing result, reqID: %q, hash: %q", exportTxSignReq.SignReqID, exportTxSignReq.Hash)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *Signer) SignImportTx(ctx context.Context, importTxHash []byte) ([65]byte, error) {
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

	if res == nil || res.Result == "" {
		return [65]byte{}, errors.Errorf("importTx got no signing result, reqID: %q, hash: %q", importTxSignReq.SignReqID, importTxSignReq.Hash)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *Signer) SignAddDelegatorTx(ctx context.Context, addDelegatorTxHash []byte) ([65]byte, error) {
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

	if res == nil || res.Result == "" {
		return [65]byte{}, errors.Errorf("addDelegatorTx got no signing result, reqID: %q, hash: %q", addDelegatorTxSignReq.SignReqID, addDelegatorTxSignReq.Hash)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}
