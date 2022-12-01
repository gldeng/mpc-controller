package keygen

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"time"
)

var (
	_ core.Task = (*RequestAdded)(nil)
)

type RequestAdded struct {
	Status Status

	Event         contract.MpcManagerKeygenRequestAdded
	KeygenRequest *types.KeygenRequest
	TxHash        *common.Hash

	group *types.Group

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
	timeOut := 10 * time.Minute
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
			if time.Now().Sub(startTime) >= timeOut {
				return nil, errors.New("task timed out")
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

		t.KeygenRequest = &types.KeygenRequest{
			ReqID:                  t.GetId(),
			CompressedPartiPubKeys: normalized,
			Threshold:              t.group.ParticipantID().Threshold(),
		}
		err = ctx.GetMpcClient().Keygen(context.Background(), t.KeygenRequest)
		if err != nil {
			return t.failIfErrorf(err, "failed to send keygen request")
		}
		prom.MpcKeygenPosted.Inc()
		t.Status = StatusKeygenReqSent
	case StatusKeygenReqSent:
		res, err := ctx.GetMpcClient().Result(context.Background(), t.KeygenRequest.ReqID)
		// TODO: Handle 404
		if err != nil {
			return t.failIfErrorf(err, "failed to get keygen result")
		}

		if res == nil || res.Result == "" {
			return t.failIfErrorf(err, "empty keygen result")
		}

		if res.Status != types.StatusDone {
			ctx.GetLogger().Debug("Keygen not done")
			return nil
		}

		prom.MpcKeygenDone.Inc()
		genPubKeyHex := res.Result
		decompressedPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(genPubKeyHex) // for Ethereum compatibility
		if err != nil {
			return t.failIfErrorf(err, "failed to decompress generated public key %v", genPubKeyHex)
		}

		txHash, err := ctx.ReportGeneratedKey(ctx.GetMyTransactSigner(), t.group.ParticipantID(), decompressedPubKeyBytes)
		if err != nil {
			return t.failIfErrorf(err, "failed to report GeneratedKey")
		}
		t.TxHash = txHash
		t.Status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckEthTx(*t.TxHash)
		ctx.GetLogger().Debugf("id %v ReportGeneratedKey Status is %v", t.GetId(), status)
		if err != nil {
			return t.failIfErrorf(err, "failed to check status for tx %x", *t.TxHash)
		}

		switch status {
		case core.TxStatusUnknown:
			return t.failIfErrorf(errors.Errorf("unkonw tx status (%v:%x) of reporting generated key for group %x",
				status, *t.TxHash, t.group.GroupId), "")
		case core.TxStatusAborted:
			t.Status = StatusKeygenReqSent // TODO: avoid endless repeating ReportGenerateKey?
			return errors.Errorf(fmt.Sprintf("ReportGeneratedKey tx %x aborted for group %x", *t.TxHash, t.group.GroupId))
		case core.TxStatusCommitted:
			t.Status = StatusDone
			ctx.GetLogger().Debugf("Generated key reported for group %x", t.group.GroupId)
		}
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
			ctx.GetLogger().Error(ErrMsgFailedToLoadGroup)
			return t.failIfErrorf(err, ErrMsgFailedToLoadGroup)
		}

		ctx.GetLogger().Debugf("Loaded group %x", group.GroupId)
		t.group = group
	}
	return nil
}
