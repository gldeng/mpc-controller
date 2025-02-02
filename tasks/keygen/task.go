package keygen

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/mpc"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	utilstime "github.com/avalido/mpc-controller/utils/time"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	_ core.Task = (*RequestAdded)(nil)
)

type RequestAdded struct {
	Status Status

	Event         contract.MpcManagerKeygenRequestAdded
	KeygenRequest *mpc.KeygenRequest
	TxHash        *common.Hash

	group *types.Group

	keygenResult string

	Failed bool
}

func NewRequestAdded(event contract.MpcManagerKeygenRequestAdded) *RequestAdded {
	return &RequestAdded{
		Status:        StatusInit,
		Event:         event,
		KeygenRequest: nil,
		TxHash:        nil,
		Failed:        false,
	}
}

func (t *RequestAdded) GetId() string {
	return fmt.Sprintf("KeygenRequestAdded(%v,%v)", t.Event.Raw.TxHash, t.Event.Raw.TxIndex)
}

func (t *RequestAdded) Next(ctx core.TaskContext) ([]core.Task, error) {
	t.retrieveGroup(ctx)
	// TODO: improve retry strategy, involving further error handling
	startTime := time.Now()
	timeout := 60 * time.Minute
	interval := 1 * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop() // TODO: stop all other timers to avoid resource leak

	var err error

loop:
	for {
		select {
		case <-timer.C:
			err = t.run(ctx)
			if t.Failed || t.Status == StatusDone {
				break loop
			}
			if time.Now().Sub(startTime) >= timeout {
				prom.TaskTimeout.With(prometheus.Labels{"flow": "", "task": "keygen"}).Inc()
				return nil, errors.New("task timeout")
			}

			timer.Reset(interval)
		}
	}

	return nil, err
}

func (t *RequestAdded) IsDone() bool {
	return t.Status == StatusDone
}

func (t *RequestAdded) FailedPermanently() bool {
	return t.Failed
}

func (t *RequestAdded) IsSequential() bool {
	return true
}

func (t *RequestAdded) run(ctx core.TaskContext) error {
	switch t.Status {
	case StatusInit:
		var pubKeys types.PubKeys
		for _, publicKey := range t.group.MemberPublicKeys {
			pubKeys = append(pubKeys, publicKey)
		}
		normalized, err := pubKeys.CompressPubKeyHexs() // for mpc-server compatibility
		if err != nil {
			return t.failIfErrorf(err, "failed to compress participant public keys")
		}

		kg := &mpc.KeygenRequest{}
		kg.RequestId = t.GetId()
		kg.ParticipantPublicKeys = normalized
		kg.Threshold = uint32(t.group.ParticipantID().Threshold())
		t.KeygenRequest = kg
		_, err = ctx.GetMpcClient().Keygen(context.Background(), t.KeygenRequest)
		if err != nil {
			return t.failIfErrorf(err, "failed to send keygen request")
		}
		prom.MpcKeygenPosted.Inc()
		t.Status = StatusKeygenReqSent
	case StatusKeygenReqSent:
		res, err := ctx.GetMpcClient().CheckResult(context.Background(), &mpc.CheckResultRequest{RequestId: t.KeygenRequest.RequestId})
		// TODO: Handle 404
		if err != nil {
			return t.failIfErrorf(err, "failed to get keygen result")
		}

		if res == nil || res.Result == "" {
			return t.failIfErrorf(err, "empty keygen result")
		}

		if res.RequestStatus != mpc.CheckResultResponse_DONE {
			ctx.GetLogger().Debug("keygen not done")
			return nil
		}

		t.keygenResult = res.Result
		t.Status = StatusKeygenReqDone
		prom.MpcKeygenDone.Inc()
		ctx.GetLogger().Info("keygen done", []logger.Field{{"generatedKey", res.Result}}...)
	case StatusKeygenReqDone:
		decompressedPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(t.keygenResult) // for Ethereum compatibility
		if err != nil {
			return t.failIfErrorf(err, "failed to decompress generated public key %v", t.keygenResult)
		}

		// waits for arbitrary duration to elapse to reduce race condition.
		utilstime.RandomDelay(5000)

		txHash, err := ctx.ReportGeneratedKey(ctx.GetMyTransactSigner(), t.group.ParticipantID(), decompressedPubKeyBytes)
		if err != nil {
			var errCreateTransactor *taskcontext.ErrTypContractBindFail
			var errExecutionReverted *taskcontext.ErrTypTxReverted
			if errors.As(err, &errCreateTransactor) || errors.As(err, &errExecutionReverted) {
				ctx.GetLogger().Error(ErrMsgReportGenPubKey, []logger.Field{{"error", err.Error()}}...)
				return t.failIfErrorf(err, ErrMsgReportGenPubKey)
			}
			ctx.GetLogger().Debug(ErrMsgReportGenPubKey, []logger.Field{{"error", err.Error()}}...)
			return errors.Wrap(err, ErrMsgReportGenPubKey)
		}
		t.TxHash = txHash
		t.Status = StatusTxSent
		ctx.GetLogger().Info("reported key", []logger.Field{{"reportedKey", bytes.BytesToHex(decompressedPubKeyBytes)}}...)
	case StatusTxSent:
		_, err := ctx.CheckEthTx(*t.TxHash)
		if err != nil {
			ctx.GetLogger().Error(ErrMsgCheckTxStatus, []logger.Field{{"tx", t.TxHash.Hex()},
				{"group", fmt.Sprintf("%x", t.group.GroupId)},
				{"error", err.Error()}}...)
			if errors.Is(err, taskcontext.ErrTxStatusAborted) {
				ctx.GetLogger().Debug("tx aborted", []logger.Field{{"tx", t.TxHash.Hex()},
					{"group", fmt.Sprintf("%x", t.group.GroupId)},
					{"error", err.Error()}}...)
				t.Status = StatusKeygenReqDone
			}
			return errors.Wrapf(err, ErrMsgCheckTxStatus)
		}

		t.Status = StatusDone
		ctx.GetLogger().Debug("key reported", []logger.Field{{"group", fmt.Sprintf("%x", t.group.GroupId)}}...)
	}
	return nil
}

func (t *RequestAdded) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *RequestAdded) retrieveGroup(ctx core.TaskContext) error {
	if t.group == nil {
		group, err := ctx.LoadGroup(t.Event.GroupId)
		if err != nil {
			ctx.GetLogger().Error(ErrMsgLoadGroup)
			return t.failIfErrorf(err, ErrMsgLoadGroup)
		}

		ctx.GetLogger().Debugf("Loaded group %x", group.GroupId)
		t.group = group
	}
	return nil
}
