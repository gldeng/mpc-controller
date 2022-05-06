package usecases

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swaggest/usecase"
)

var globalSiner crypto.Signer

func Keygen() usecase.IOInteractor {
	u := usecase.NewIOI(new(KeygenInput), nil, func(ctx context.Context, input, output interface{}) error {
		var (
			in = input.(*KeygenInput)
		)

		lastKeygenReq := storer.GetKeygenRequestModel(in.RequestId)
		if lastKeygenReq == nil {
			lastKeygenReq = &KeygenRequestModel{
				input:   in,
				reqType: TypeKeygen,
				hits:    1,
				status:  StatusReceived,
			}
			storer.StoreKeygenRequestModel(lastKeygenReq)
			logger.Debug("Mpc-server received keygen request",
				logger.Field{"reqId", in.RequestId},
				logger.Field{"hits", lastKeygenReq.hits},
				logger.Field{"status", lastKeygenReq.status},
				logger.Field{"pubkey", lastKeygenReq.result})
			return nil
		}

		if lastKeygenReq.hits != 2 {
			lastKeygenReq.hits++
			storer.StoreKeygenRequestModel(lastKeygenReq)
			logger.Debug("Mpc-server received keygen request",
				logger.Field{"reqId", in.RequestId},
				logger.Field{"hits", lastKeygenReq.hits},
				logger.Field{"status", lastKeygenReq.status},
				logger.Field{"pubkey", lastKeygenReq.result})
			return nil
		}

		lastKeygenReq.hits++
		signer, _ := crypto.NewSECP256K1RSigner()
		pubkeyHex := common.Bytes2Hex(signer.PublicKey().Bytes())
		lastKeygenReq.signer = signer
		lastKeygenReq.result = pubkeyHex
		lastKeygenReq.status = StatusDone
		storer.StoreKeygenRequestModel(lastKeygenReq)
		logger.Debug("Mpc-server received keygen request",
			logger.Field{"reqId", in.RequestId},
			logger.Field{"hits", lastKeygenReq.hits},
			logger.Field{"status", lastKeygenReq.status},
			logger.Field{"pubkey", lastKeygenReq.result})

		globalSiner = signer

		return nil
	})

	u.SetTitle("Generate key")

	return u
}
