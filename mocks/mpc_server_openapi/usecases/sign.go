package usecases

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/swaggest/usecase"
	"strings"
)

func Sign() usecase.IOInteractor {
	u := usecase.NewIOI(new(SignInput), nil, func(ctx context.Context, input, output interface{}) error {
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
			logger.Debug("Mpc-server received sign request",
				logger.Field{"reqId", in.RequestId},
				logger.Field{"hits", lastSignReq.hits},
				logger.Field{"status", lastSignReq.status},
				logger.Field{"signature", lastSignReq.result})
			return nil
		}

		if lastSignReq.hits != 1 {
			lastSignReq.hits++
			storer.StoreSignRequestModel(lastSignReq)
			logger.Debug("Mpc-server received sign request",
				logger.Field{"reqId", in.RequestId},
				logger.Field{"hits", lastSignReq.hits},
				logger.Field{"status", lastSignReq.status},
				logger.Field{"signature", lastSignReq.result})
			return nil
		}

		reqIdParts := strings.Split(in.RequestId, "-")

		lastKeygenReq := storer.GetKeygenRequestModel(reqIdParts[0])
		if lastKeygenReq == nil || lastKeygenReq.status != StatusDone {
			logger.Error("Mpc-server failed to get key to sign",
				logger.Field{"reqId", in.RequestId})
			return errors.Errorf("Mpc-server failed to get key to sign, request id: %v", in.RequestId)
		}

		lastSignReq.hits++
		signer := lastKeygenReq.signer
		digest := common.Hex2Bytes(lastSignReq.input.Hash)

		sigBytes, err := signer.SignHash(digest)
		if err != nil {
			logger.Error("Mpc-server failed to sign",
				logger.Field{"reqId", in.RequestId},
				logger.Field{"error", err})
			return errors.Wrapf(err, "Mpc-server failed to sign")
		}
		sigHex := common.Bytes2Hex(sigBytes)
		lastSignReq.result = sigHex
		lastSignReq.status = StatusDone
		storer.StoreSignRequestModel(lastSignReq)
		logger.Debug("Mpc-server received sign request",
			logger.Field{"reqId", in.RequestId},
			logger.Field{"hits", lastSignReq.hits},
			logger.Field{"status", lastSignReq.status},
			logger.Field{"signature", lastSignReq.result})
		return nil
	})

	u.SetTitle("Sign digest in hex format")

	return u
}
