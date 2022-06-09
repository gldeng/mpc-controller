package staking

import (
	"context"
	ctlPk "github.com/avalido/mpc-controller"
	"github.com/avalido/mpc-controller/contract"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

type OnStakeRequestAddedTask struct {
	PubKeyHashHex string
	Signer        *bind.TransactOpts
	CheckRcptDur  time.Duration // duration to check transaction to see whether it is successful
	ctlPk.Manager
}

// todo: introduce effective strategy to deal with failed transaction

func (s *OnStakeRequestAddedTask) Do(ctx context.Context, evt ctlPk.EvtFromContractStakeRequestAdded) {
	// Join request
	err := s.joinRequest(ctx, evt.Evt)
	if err != nil {
		err = errors.Wrapf(err, "failed to join request")
		evt.ReplyCh <- struct{ Error error }{Error: err}
		return
	}
}

func (s *OnStakeRequestAddedTask) joinRequest(ctx context.Context, req *contract.MpcManagerStakeRequestAdded) error {
	// Get participant index in a group
	index, err := s.reqGetPartiIndex(ctx, req.PublicKey.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	// Join request
	txHash, err := s.reqJoinRequest(ctx, req.RequestId, index)
	if err != nil {
		return errors.WithStack(err)
	}

	time.Sleep(s.CheckRcptDur)
	err = s.reqTxReceipt(ctx, *txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *OnStakeRequestAddedTask) reqGetPartiIndex(ctx context.Context, pubKeyHex string) (*big.Int, error) {
	var reqIndex = ctlPk.ReqToStorerGetParticipantIndex{
		PartiPubKeyHashHex: s.PubKeyHashHex,
		GenPubKeyHexHex:    pubKeyHex,
		ReplyCh: make(chan struct {
			Index *big.Int
			Error error
		}),
	}
	err := s.Request(ctx, &reqIndex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reply := <-reqIndex.ReplyCh
	if reply.Error != nil {
		return nil, errors.WithStack(reply.Error)
	}

	return reply.Index, nil
}

func (s *OnStakeRequestAddedTask) reqJoinRequest(ctx context.Context, reqId, index *big.Int) (*common.Hash, error) {
	var joinReq = ctlPk.ReqToContractJoinRequest{
		Signer: s.Signer,
		ReqId:  reqId,
		Index:  index,
		ReplyCh: make(chan struct {
			Tx    *types.Transaction
			Error error
		}),
	}

	err := s.Request(ctx, &joinReq)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reply := <-joinReq.ReplyCh
	if reply.Error != nil {
		return nil, errors.WithStack(reply.Error)
	}

	txHash := reply.Tx.Hash()
	return &txHash, nil
}

func (s *OnStakeRequestAddedTask) reqTxReceipt(ctx context.Context, txHash common.Hash) error {
	var txRcpt = ctlPk.ReqToCChainTransactionReceipt{
		TxHash: txHash,
		ReplyCh: make(chan struct {
			Receipt *types.Receipt
			Error   error
		}),
	}

	err := s.Request(ctx, &txRcpt)
	if err != nil {
		return errors.WithStack(err)
	}

	reply := <-txRcpt.ReplyCh
	if reply.Error != nil {
		return errors.WithStack(reply.Error)
	}

	if reply.Receipt.Status != 1 {
		return errors.New("Transaction failed")
	}

	return nil
}
