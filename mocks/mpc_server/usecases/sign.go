package usecases

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/swaggest/usecase"
	"sync"
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
				{"signature", lastSignReq.result}}...)
			return nil
		}

		if lastSignReq.hits != 1 {
			lastSignReq.hits++
			storer.StoreSignRequestModel(lastSignReq)
			logger.Debug("Mpc-server received sign request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastSignReq.hits},
				{"status", lastSignReq.status},
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
			{"signature", lastSignReq.result}}...)
		return nil
	})

	u.SetTitle("Sign digest in hex format")

	return u
}
