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
	CompressedGenPubKey    string
}

func (s *Signer) SignExportTx(ctx context.Context, exportTxHash []byte) ([65]byte, error) {
	exportTxSignReq := core.SignRequest{
		RequestId:              s.TaskID + "-" + strconv.Itoa(0),
		CompressedGenPubKey:    s.CompressedGenPubKey,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	res, err := s.SignDone(ctx, &exportTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export tx, RequestID: %q", exportTxSignReq.RequestId)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *Signer) SignImportTx(ctx context.Context, importTxHash []byte) ([65]byte, error) {
	importTxSignReq := core.SignRequest{
		RequestId:              s.TaskID + "-" + strconv.Itoa(1),
		CompressedGenPubKey:    s.CompressedGenPubKey,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	res, err := s.SignDone(ctx, &importTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export RequestID: %q", importTxSignReq.RequestId)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}

func (s *Signer) SignAddDelegatorTx(ctx context.Context, addDelegatorTxHash []byte) ([65]byte, error) {
	addDelegatorTxSignReq := core.SignRequest{
		RequestId:              s.TaskID + "-" + strconv.Itoa(2),
		CompressedGenPubKey:    s.CompressedGenPubKey,
		CompressedPartiPubKeys: s.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(addDelegatorTxHash),
	}

	res, err := s.SignDone(ctx, &addDelegatorTxSignReq)
	if err != nil {
		return [65]byte{}, errors.Wrapf(err, "failed to sign export RequestID: %q", addDelegatorTxSignReq.RequestId)
	}

	return bytes.BytesTo65Bytes(bytes.HexToBytes(res.Result)), nil
}
