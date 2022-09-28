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
		ID:                     s.TaskID + "-" + strconv.Itoa(0),
		CompressedGenPubKeyHex: s.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	res, err := s.SignDone(ctx, &exportTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export tx, ReqNo: %q", exportTxSignReq.ID)
	}

	if res == nil || res.Result == "" {
		return [65]byte{}, errors.Errorf("exportTx got no signing result, reqID: %q, hash: %q", exportTxSignReq.ID, exportTxSignReq.Hash)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *Signer) SignImportTx(ctx context.Context, importTxHash []byte) ([65]byte, error) {
	importTxSignReq := core.SignRequest{
		ID:                     s.TaskID + "-" + strconv.Itoa(1),
		CompressedGenPubKeyHex: s.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	res, err := s.SignDone(ctx, &importTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export ReqNo: %q", importTxSignReq.ID)
	}

	if res == nil || res.Result == "" {
		return [65]byte{}, errors.Errorf("importTx got no signing result, reqID: %q, hash: %q", importTxSignReq.ID, importTxSignReq.Hash)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *Signer) SignAddDelegatorTx(ctx context.Context, addDelegatorTxHash []byte) ([65]byte, error) {
	addDelegatorTxSignReq := core.SignRequest{
		ID:                     s.TaskID + "-" + strconv.Itoa(2),
		CompressedGenPubKeyHex: s.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(addDelegatorTxHash),
	}

	res, err := s.SignDone(ctx, &addDelegatorTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export ReqNo: %q", addDelegatorTxSignReq.ID)
	}

	if res == nil || res.Result == "" {
		return [65]byte{}, errors.Errorf("addDelegatorTx got no signing result, reqID: %q, hash: %q", addDelegatorTxSignReq.ID, addDelegatorTxSignReq.Hash)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}
