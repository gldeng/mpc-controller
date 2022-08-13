package transactor

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	bind2 "github.com/avalido/mpc-controller/utils/contract/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var _ Transactor = (*MyTransactor)(nil)

const (
	MethodJoinRequest        Method = "joinRequest"
	MethodReportGeneratedKey Method = "reportGeneratedKey"
)

type Method string

type Transactor interface {
	JoinRequest(ctx context.Context, partiId [32]byte, reqHash [32]byte) (*types.Transaction, *types.Receipt, error)
	ReportGeneratedKey(ctx context.Context, partiId [32]byte, genPubKey []byte) (*types.Transaction, *types.Receipt, error)
	RetryTransact(ctx context.Context, fn Transact) (*types.Transaction, *types.Receipt, error)
}

type Transact func() (tx *types.Transaction, err error, retry bool, checkReceipt bool)

type MyTransactor struct {
	Auth               *bind.TransactOpts
	ContractAddr       common.Address
	ContractTransactor bind.ContractTransactor
	EthClient          chain.EthClient
	Logger             logger.Logger
	bind2.BoundTransactor
}

func (t *MyTransactor) Init(ctx context.Context) {
	boundFilterer, err := contract.BindMpcManagerTransactor(t.ContractAddr, t.ContractTransactor)
	t.Logger.FatalOnError(err, "Failed to bind MpcManager filterer")
	t.BoundTransactor = boundFilterer
}

func (t *MyTransactor) JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, *types.Receipt, error) {
	fn := func() (tx *types.Transaction, err error, retry bool, checkReceipt bool) {
		tx, err = t.BoundTransactor.Transact(t.Auth, string(MethodJoinRequest), participantId, requestHash)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "QuorumAlreadyReached"):
				return tx, errors.WithStack(&ErrTypQuorumAlreadyReached{ErrMsg: errMsg, Cause: err}), false, false
			case strings.Contains(errMsg, "AttemptToRejoin "):
				return tx, errors.WithStack(&ErrTypAttemptToRejoin{ErrMsg: errMsg, Cause: err}), false, false
			}
			return tx, errors.WithStack(err), true, true
		}
		return tx, nil, false, true
	}

	tx, rcpt, err := t.RetryTransact(ctx, fn)
	partiIdHex := bytes.Bytes32ToHex(participantId)
	reqHashHex := bytes.Bytes32ToHex(requestHash)
	return tx, rcpt, errors.Wrapf(err, "failed to join request. partiId:%v, reqHash:%v", partiIdHex, reqHashHex)
}

func (t *MyTransactor) ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, *types.Receipt, error) {
	fn := func() (tx *types.Transaction, err error, retry bool, checkReceipt bool) {
		tx, err = t.BoundTransactor.Transact(t.Auth, string(MethodReportGeneratedKey), participantId, generatedPublicKey)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "AttemptToReconfirmKey"):
				return tx, errors.WithStack(&ErrTypAttemptToReconfirmKey{ErrMsg: errMsg, Cause: err}), false, false
			}
			return tx, errors.WithStack(err), true, true
		}
		return tx, nil, false, true
	}

	tx, rcpt, err := t.RetryTransact(ctx, fn)
	partiIdHex := bytes.Bytes32ToHex(participantId)
	genPubKeyHex := bytes.BytesToHex(generatedPublicKey)
	return tx, rcpt, errors.Wrapf(err, "failed to report generated public key. partiId:%v, genPubKey:%v", partiIdHex, genPubKeyHex)
}

func (t *MyTransactor) RetryTransact(ctx context.Context, fn Transact) (*types.Transaction, *types.Receipt, error) {
	var tx *types.Transaction
	var rcpt *types.Receipt

	err := backoff.RetryRetryFn10Times(ctx, func() (retry bool, err error) {
		var checkReceipt bool
		tx, err, retry, checkReceipt = fn()
		if retry {
			return true, errors.WithStack(err)
		}

		if !checkReceipt {
			return false, errors.WithStack(err)
		}

		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (retry bool, err error) {
			rcpt, err = t.EthClient.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				return true, errors.WithStack(err)
			}

			if rcpt.Status != 1 {
				return true, errors.WithStack(&ErrTypTransactionFailed{})
			}

			return false, nil
		})
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})

	if err != nil {
		return tx, nil, errors.WithStack(err)
	}
	return tx, rcpt, errors.WithStack(err)
}
