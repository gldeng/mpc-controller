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
	Transact(ctx context.Context, fn TransactFn) (*types.Transaction, *types.Receipt, error)
}

type TransactFn func() (tx *types.Transaction, err error, retry bool)

type MyTransactor struct {
	Logger             logger.Logger
	Auth               *bind.TransactOpts
	ContractAddr       common.Address
	Receipter          chain.Receipter
	ContractTransactor bind.ContractTransactor
	boundTransactor    bind2.BoundTransactor
}

func (t *MyTransactor) Init(ctx context.Context) {
	boundFilterer, err := contract.BindMpcManagerTransactor(t.ContractAddr, t.ContractTransactor)
	t.Logger.FatalOnError(err, "Failed to bind MpcManager filterer")
	t.boundTransactor = boundFilterer
}

func (t *MyTransactor) JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, *types.Receipt, error) {
	fn := func() (tx *types.Transaction, err error, retry bool) {
		tx, err = t.boundTransactor.Transact(t.Auth, string(MethodJoinRequest), participantId, requestHash)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "QuorumAlreadyReached"):
				return tx, errors.WithStack(&ErrTypQuorumAlreadyReached{Cause: err}), false
			case strings.Contains(errMsg, "AttemptToRejoin "):
				return tx, errors.WithStack(&ErrTypAttemptToRejoin{Cause: err}), false
			}
			return tx, errors.WithStack(err), true
		}
		return tx, nil, false
	}

	tx, rcpt, err := t.Transact(ctx, fn)
	partiIdHex := bytes.Bytes32ToHex(participantId)
	reqHashHex := bytes.Bytes32ToHex(requestHash)
	return tx, rcpt, errors.Wrapf(err, "failed to join request. partiId:%v, reqHash:%v", partiIdHex, reqHashHex)
}

func (t *MyTransactor) ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, *types.Receipt, error) {
	fn := func() (tx *types.Transaction, err error, retry bool) {
		tx, err = t.boundTransactor.Transact(t.Auth, string(MethodReportGeneratedKey), participantId, generatedPublicKey)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "AttemptToReconfirmKey"):
				return tx, errors.WithStack(&ErrTypAttemptToReconfirmKey{Cause: err}), false
			}
			return tx, errors.WithStack(err), true
		}
		return tx, nil, false
	}

	tx, rcpt, err := t.Transact(ctx, fn)
	partiIdHex := bytes.Bytes32ToHex(participantId)
	genPubKeyHex := bytes.BytesToHex(generatedPublicKey)
	return tx, rcpt, errors.Wrapf(err, "failed to report generated public key. partiId:%v, genPubKey:%v", partiIdHex, genPubKeyHex)
}

func (t *MyTransactor) Transact(ctx context.Context, fn TransactFn) (*types.Transaction, *types.Receipt, error) {
	var tx *types.Transaction
	err := backoff.RetryRetryFnForever(ctx, func() (retry bool, err error) {
		tx, err, retry = fn()
		if retry {
			return true, errors.WithStack(err)
		}
		return false, errors.WithStack(err)
	})

	if err != nil {
		return tx, nil, errors.WithStack(err)
	}

	var rcpt *types.Receipt
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		rcpt, err = t.Receipter.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return true, errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			return true, errors.WithStack(&ErrTypTransactionFailed{})
		}

		return false, nil
	})

	return tx, rcpt, errors.WithStack(err)
}

//func (t *MyTransactor) JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, *types.Receipt, error) {
//	var tx *types.Transaction
//	err := backoff.RetryRetryFnForever(ctx, func() (retry bool, err error) {
//		tx, err = t.boundTransactor.Transact(t.Auth, string(MethodJoinRequest), participantId, requestHash)
//		if err != nil {
//			errMsg := err.Error()
//			switch {
//			case strings.Contains(errMsg, "QuorumAlreadyReached"):
//				return false, errors.WithStack(&ErrTypQuorumAlreadyReached{Cause: err})
//			case strings.Contains(errMsg, "AttemptToRejoin "):
//				return false, errors.WithStack(&ErrTypAttemptToRejoin{Cause: err})
//			}
//			return true, errors.WithStack(err)
//		}
//		return false, nil
//	})
//
//	partiIdHex := bytes.Bytes32ToHex(participantId)
//	reqHashHex := bytes.Bytes32ToHex(requestHash)
//	err = errors.Wrapf(err, "failed to join request. partiId:%v, reqHash:%v", partiIdHex, reqHashHex)
//	if err != nil {
//		return tx, nil, err
//	}
//
//	var rcpt *types.Receipt
//	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
//		rcpt, err = t.Receipter.TransactionReceipt(ctx, tx.Hash())
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//
//		if rcpt.Status != 1 {
//			return true, errors.WithStack(&ErrTypTransactionFailed{})
//		}
//
//		return false, nil
//	})
//
//	err = errors.Wrapf(err, "failed to join request. partiId:%v, reqHash:%v", partiIdHex, reqHashHex)
//	if err != nil {
//		return nil, rcpt, errors.WithStack(err)
//	}
//
//	return tx, rcpt, nil
//}

//func (t *MyTransactor) ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, *types.Receipt, error) {
//	var tx *types.Transaction
//	err := backoff.RetryRetryFnForever(ctx, func() (retry bool, err error) {
//		tx, err = t.boundTransactor.Transact(t.Auth, string(MethodReportGeneratedKey), participantId, generatedPublicKey)
//		if err != nil {
//			errMsg := err.Error()
//			switch {
//			case strings.Contains(errMsg, "AttemptToReconfirmKey"):
//				return false, errors.WithStack(&ErrTypAttemptToReconfirmKey{Cause: err})
//			}
//			return true, errors.WithStack(err)
//		}
//		return false, nil
//	})
//
//	partiIdHex := bytes.Bytes32ToHex(participantId)
//	genPubKeyHex := bytes.BytesToHex(generatedPublicKey)
//	err = errors.Wrapf(err, "failed to report generated public key. partiId:%v, genPubKey:%v", partiIdHex, genPubKeyHex)
//	if err != nil {
//		return tx, nil, err
//	}
//
//	var rcpt *types.Receipt
//	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
//		rcpt, err = t.Receipter.TransactionReceipt(ctx, tx.Hash())
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//
//		if rcpt.Status != 1 {
//			return true, errors.WithStack(&ErrTypTransactionFailed{})
//		}
//
//		return false, nil
//	})
//
//	err = errors.Wrapf(err, "failed to report generated public key. partiId:%v, genPubKey:%v", partiIdHex, genPubKeyHex)
//	if err != nil {
//		return nil, rcpt, errors.WithStack(err)
//	}
//
//	return tx, rcpt, nil
//}
