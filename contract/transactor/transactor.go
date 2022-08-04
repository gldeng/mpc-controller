package transactor

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	bind2 "github.com/avalido/mpc-controller/utils/contract/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"time"
)

const (
	MethodJoinRequest        Method = "joinRequest"
	MethodReportGeneratedKey Method = "reportGeneratedKey"
)

var (
	ErrCannotJoin = errors.New("Cannot join anymore")
)

type Method string

type Transactor interface {
	JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, error)
	ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, error)
}

type MyTransactor struct {
	Logger             logger.Logger
	ContractAddr       common.Address
	ContractTransactor bind.ContractTransactor
	Auth               *bind.TransactOpts
	boundTransactor    bind2.BoundTransactor
}

func (t *MyTransactor) Init(ctx context.Context) {
	boundFilterer, err := contract.BindMpcManagerTransactor(t.ContractAddr, t.ContractTransactor)
	t.Logger.FatalOnError(err, "Failed to bind MpcManager filterer")
	t.boundTransactor = boundFilterer
}

func (t *MyTransactor) JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, error) {
	var tx *types.Transaction
	err := backoff.RetryRetryFn(ctx, backoff.ConstantPolicyForever(time.Second*60), backoff.ExponentialPolicy10Times(time.Second, time.Second*10), func() (retry bool, err error) {
		tx, err = t.boundTransactor.Transact(t.Auth, string(MethodJoinRequest), participantId, requestHash)
		if err != nil {
			if strings.Contains(err.Error(), "Cannot join anymore") { // todo: improve error handling
				return false, errors.Wrapf(ErrCannotJoin, "failed to join request, participantId:%v, requestHash:%v", bytes.Bytes32ToHex(participantId), bytes.Bytes32ToHex(requestHash))
			}
			return true, errors.WithStack(err)
		}
		return false, nil
	})

	err = errors.Wrapf(err, "failed to join request")
	if err != nil {
		return nil, err
	}

}

func (t *MyTransactor) ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, error) {
	return t.boundTransactor.Transact(t.Auth, string(MethodReportGeneratedKey), participantId, generatedPublicKey)
}

func (eh *StakeRequestAddedEventHandler) doJoinRequest(ctx context.Context, myIndex *big.Int, req *contract.MpcManagerStakeRequestAdded) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
			var err error
			tx, err = transactor.JoinRequest(eh.Signer, req.RequestId, myIndex)
			if err != nil {
				if strings.Contains(err.Error(), "Cannot join anymore") {
					tx = nil
					return false, errors.WithStack(ErrCannotJoin)
				}
				return true, errors.WithStack(err)
			}
			return false, nil
		})

		err = errors.Wrapf(err, "failed to join request")
		if err != nil {
			if errors.Is(err, ErrCannotJoin) {
				return false, err
			}
			return true, err
		}

		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
			rcpt, err := eh.Receipter.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				return true, errors.WithStack(err)
			}

			if rcpt.Status != 1 {
				return true, errors.New("join request tx receipt status != 1")
			}

			return false, nil
		})
		if err != nil {
			return true, errors.WithStack(err)
		}

		return false, nil
	})

	err = errors.Wrapf(err, "failed to join request. reqID:%v, partiIndex:%v", req.RequestId, myIndex)
	if err != nil {
		return
	}
	if tx != nil {
		txHash_ := tx.Hash()
		txHash = &txHash_
		return

	}
	return
}
