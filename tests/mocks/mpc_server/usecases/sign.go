package usecases

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/swaggest/usecase"
	"sync"
)

const (
	ErrMsgSignReqRefused = "sign request refused"
)

var lockSign = &sync.Mutex{}

func Sign() usecase.IOInteractor {
	u := usecase.NewIOI(new(SignInput), nil, func(ctx context.Context, input, output interface{}) error {
		lockSign.Lock()
		defer lockSign.Unlock()

		var (
			in = input.(*SignInput)
		)

		lastSignReq := storer.GetSignRequestModel(in.RequestId)
		if lastSignReq == nil {
			lastSignReq = &SignRequestModel{
				input:   in,
				reqType: TypeSign,
				hits:    1,
				status:  StatusReceived,
			}
			storer.StoreSignRequestModel(lastSignReq)
			logger.Debug("Mpc-server received sign request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastSignReq.hits},
				{"status", lastSignReq.status},
				{"hash", in.Hash},
				{"signature", lastSignReq.result}}...)
			return nil
		}

		if in.PublicKey != lastSignReq.input.PublicKey {
			err := errors.Errorf("Inconsistent public key for sign request %q, expected public key %q , but received %q", in.RequestId, lastSignReq.input.PublicKey, in.PublicKey)
			err = errors.Wrap(err, ErrMsgSignReqRefused)
			logger.ErrorOnError(err, ErrMsgSignReqRefused)
			lastSignReq.status = StatusError + ": " + RequestStatus(err.Error())
			return err
		}

		if len(in.ParticipantKeys) != len(lastSignReq.input.ParticipantKeys) {
			err := errors.Errorf("Inconsistent participants length for sign request %q, expected participants length %v , but received %v", in.RequestId, len(lastSignReq.input.ParticipantKeys), len(in.ParticipantKeys))
			err = errors.Wrap(err, ErrMsgSignReqRefused)
			logger.ErrorOnError(err, ErrMsgSignReqRefused)
			lastSignReq.status = StatusError + ": " + RequestStatus(err.Error())
			return err
		}

		for i, partiKey := range in.ParticipantKeys {
			if partiKey != lastSignReq.input.ParticipantKeys[i] {
				err := errors.Errorf("Inconsistent participant public key at index %v for sign request %q, expected participant key %q , but received %q", i, in.RequestId, lastSignReq.input.ParticipantKeys[i], partiKey)
				err = errors.Wrap(err, ErrMsgSignReqRefused)
				logger.ErrorOnError(err, ErrMsgSignReqRefused)
				lastSignReq.status = StatusError + ": " + RequestStatus(err.Error())
				return err
			}
		}

		if in.Hash != lastSignReq.input.Hash {
			err := errors.Errorf("Inconsistent hash for sign request %q, expected hash %q , but received %q", in.RequestId, lastSignReq.input.Hash, in.Hash)
			err = errors.Wrap(err, ErrMsgSignReqRefused)
			logger.ErrorOnError(err, ErrMsgSignReqRefused)
			lastSignReq.status = StatusError + ": " + RequestStatus(err.Error())
			return err
		}

		if lastSignReq.hits == Threshold+1 {
			lastSignReq.hits++
			storer.StoreSignRequestModel(lastSignReq)
			logger.Error("Received redundant sign request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastSignReq.hits},
				{"status", lastSignReq.status},
				{"hash", in.Hash},
				{"signature", lastSignReq.result}}...)
			return errors.Errorf("Sign for request %q has been done, extra request not allowed", in.RequestId)
		}

		if lastSignReq.hits != Threshold {
			lastSignReq.hits++
			storer.StoreSignRequestModel(lastSignReq)
			logger.Debug("Mpc-server received sign request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastSignReq.hits},
				{"status", lastSignReq.status},
				{"hash", in.Hash},
				{"signature", lastSignReq.result}}...)
			return nil
		}

		//reqIdParts := strings.Split(in.SignReqID, "-")

		//lastKeygenReq := storer.GetKeygenRequestModel(reqIdParts[0])
		//if lastKeygenReq == nil || lastKeygenReq.status != StatusDone {
		//	logger.Error("Mpc-server failed to get key to sign",
		//		logger.Field{"reqId", in.SignReqID})
		//	return errors.Errorf("Mpc-server failed to get key to sign, request id: %v", in.SignReqID)
		//}

		lastSignReq.hits++
		//signer := lastKeygenReq.signer
		signer := globalSiner
		digest := common.Hex2Bytes(lastSignReq.input.Hash)

		sigBytes, err := signer.SignHash(digest)
		if err != nil {
			logger.Error("Mpc-server failed to sign", []logger.Field{
				{"reqId", in.RequestId},
				{"error", err}}...)
			return errors.Wrapf(err, "Mpc-server failed to sign")
		}
		sigHex := common.Bytes2Hex(sigBytes)
		lastSignReq.result = sigHex
		lastSignReq.status = StatusDone
		storer.StoreSignRequestModel(lastSignReq)
		logger.Debug("Mpc-server received sign request", []logger.Field{
			{"reqId", in.RequestId},
			{"hits", lastSignReq.hits},
			{"status", lastSignReq.status},
			{"hash", in.Hash},
			{"signature", lastSignReq.result}}...)
		return nil
	})

	u.SetTitle("Sign digest in hex format")

	return u
}
