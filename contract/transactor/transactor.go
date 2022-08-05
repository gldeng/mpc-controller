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

const (
	MethodJoinRequest        Method = "joinRequest"
	MethodReportGeneratedKey Method = "reportGeneratedKey"
)

type Method string

type Transactor interface {
	JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, *types.Receipt, error)
	ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, *types.Receipt, error)
}

type MyTransactor struct {
	Logger             logger.Logger
	ContractAddr       common.Address
	Receipter          chain.Receipter
	ContractTransactor bind.ContractTransactor
	Auth               *bind.TransactOpts
	boundTransactor    bind2.BoundTransactor
}

func (t *MyTransactor) Init(ctx context.Context) {
	boundFilterer, err := contract.BindMpcManagerTransactor(t.ContractAddr, t.ContractTransactor)
	t.Logger.FatalOnError(err, "Failed to bind MpcManager filterer")
	t.boundTransactor = boundFilterer
}

func (t *MyTransactor) JoinRequest(ctx context.Context, participantId [32]byte, requestHash [32]byte) (*types.Transaction, *types.Receipt, error) {
	var tx *types.Transaction
	err := backoff.RetryRetryFnForever(ctx, func() (retry bool, err error) {
		tx, err = t.boundTransactor.Transact(t.Auth, string(MethodJoinRequest), participantId, requestHash)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "QuorumAlreadyReached"):
				return false, errors.WithStack(&ErrTypQuorumAlreadyReached{Cause: err})
			case strings.Contains(errMsg, "AttemptToRejoin "):
				return false, errors.WithStack(&ErrTypAttemptToRejoin{Cause: err})
			}
			return true, errors.WithStack(err)
		}
		return false, nil
	})

	participantIdHex := bytes.Bytes32ToHex(participantId)
	requestHashHex := bytes.Bytes32ToHex(requestHash)
	err = errors.Wrapf(err, "failed to join request. participantId:%v, requestHash:%v", participantIdHex, requestHashHex)
	if err != nil {
		return tx, nil, err
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

	err = errors.Wrapf(err, "failed to join request. participantId:%v, requestHash:%v", participantIdHex, requestHashHex)
	if err != nil {
		return nil, rcpt, errors.WithStack(err)
	}

	return tx, rcpt, nil
}

func (t *MyTransactor) ReportGeneratedKey(ctx context.Context, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, error) {
	return t.boundTransactor.Transact(t.Auth, string(MethodReportGeneratedKey), participantId, generatedPublicKey)
}
