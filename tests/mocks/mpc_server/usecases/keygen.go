package usecases

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/address"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/swaggest/usecase"
	"sync"
)

var globalSiner crypto.Signer_

var lockKeygen = &sync.Mutex{}

func Keygen() usecase.IOInteractor {

	u := usecase.NewIOI(new(KeygenInput), nil, func(ctx context.Context, input, output interface{}) error {
		lockKeygen.Lock()
		defer lockKeygen.Unlock()

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
			logger.Debug("Mpc-server received keygen request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastKeygenReq.hits},
				{"status", lastKeygenReq.status},
				{"pubkey", lastKeygenReq.result}}...)
			return nil
		}

		if lastKeygenReq.hits == Participants {
			lastKeygenReq.hits++
			storer.StoreKeygenRequestModel(lastKeygenReq)
			logger.Error("Received redundant keygen request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastKeygenReq.hits},
				{"status", lastKeygenReq.status},
				{"pubkey", lastKeygenReq.result}}...)
			return errors.Errorf("keygen for request %q has been done, extra request not allowed", in.RequestId)
		}

		if lastKeygenReq.hits != Participants-1 {
			lastKeygenReq.hits++
			storer.StoreKeygenRequestModel(lastKeygenReq)
			logger.Debug("Mpc-server received keygen request", []logger.Field{
				{"reqId", in.RequestId},
				{"hits", lastKeygenReq.hits},
				{"status", lastKeygenReq.status},
				{"pubkey", lastKeygenReq.result}}...)
			return nil
		}

		lastKeygenReq.hits++
		signer, _ := crypto.NewSECP256K1RSigner()
		pubkeyHex := common.Bytes2Hex(signer.PublicKey().Bytes())
		lastKeygenReq.signer = signer
		lastKeygenReq.result = pubkeyHex
		lastKeygenReq.status = StatusDone
		storer.StoreKeygenRequestModel(lastKeygenReq)
		logger.Debug("Mpc-server received keygen request", []logger.Field{
			{"reqId", in.RequestId},
			{"hits", lastKeygenReq.hits},
			{"status", lastKeygenReq.status},
			{"pubkey", lastKeygenReq.result}}...)

		globalSiner = signer
		signerKeyBytes := signer.PrivateKey().Bytes()
		signerKeyHex := common.Bytes2Hex(signerKeyBytes)
		signerPubHex := common.Bytes2Hex(signer.PublicKey().Bytes())
		cChainAddr, _ := address.PubKeyBytesToAddress(signer.PublicKey().Bytes())
		pChainAddr := signer.Address()
		logger.Info("Mpc mock server generated a signer", []logger.Field{
			{"privateKey", signerKeyHex},
			{"publicKey", signerPubHex},
			{"cChainAddr", cChainAddr},
			{"pChainAddr", pChainAddr}}...)

		return nil
	})

	u.SetTitle("Generate key")

	return u
}
