package keygen

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
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
	KeygenRequest *core.KeygenRequest
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
	group, err := ctx.LoadGroup(t.Event.GroupId)
	if err != nil {
		ctx.GetLogger().ErrorOnError(err, ErrMsgFailedToLoadGroup)
		return nil, t.failIfError(err, ErrMsgFailedToLoadGroup)
	}

	groupIDHex := bytes.Bytes32ToHex(group.GroupId)
	ctx.GetLogger().Debug(fmt.Sprintf("loaded group %v\n", groupIDHex))
	t.group = group

	interval := 100 * time.Millisecond
	timer := time.NewTimer(interval)
	defer timer.Stop() // TODO: stop all other timers to avoid resource leak

loop:
	for {
		select {
		case <-timer.C:
			err = t.run(ctx)
			if t.Failed || t.Status == StatusDone {
				break loop
			}

			timer.Reset(interval)
		}
	}

	ctx.GetLogger().ErrorOnError(err, fmt.Sprintf("Keygen error for %s", groupIDHex))
	ctx.GetLogger().DebugNilError(err, fmt.Sprintf("Keygen done for %s", groupIDHex))

	return nil, err
}

func (t *RequestAdded) IsDone() bool {
	return t.Status == StatusDone
}

func (t *RequestAdded) FailedPermanently() bool {
	return t.Failed
}

func (t *RequestAdded) RequiresNonce() bool {
	return false
}

func (t *RequestAdded) run(ctx core.TaskContext) error {
	switch t.Status {
	case StatusInit:
		var pubKeys storage.PubKeys
		for _, publicKey := range t.group.MemberPublicKeys {
			pubKeys = append(pubKeys, publicKey)
		}
		normalized, err := pubKeys.CompressPubKeyHexs() // for mpc-server compatibility
		if err != nil {
			return t.failIfError(err, "failed to compress participant public keys")
		}

		t.KeygenRequest = &core.KeygenRequest{
			ReqID:                  t.GetId(),
			CompressedPartiPubKeys: normalized,
			Threshold:              0, // TODO: fix it
		}
		err = ctx.GetMpcClient().Keygen(context.Background(), t.KeygenRequest)
		if err != nil {
			return t.failIfError(err, "failed to send keygen request")
		}
		t.Status = StatusKeygenReqSent
	case StatusKeygenReqSent:
		res, err := ctx.GetMpcClient().Result(context.Background(), t.KeygenRequest.ReqID)
		// TODO: Handle 404
		if err != nil {
			return t.failIfError(err, "failed to get result")
		}

		if res.Status != core.StatusDone {
			ctx.GetLogger().Debug("Keygen not done")
			return nil
		}
		generatedPubKey := common.Hex2Bytes(res.Result)
		txHash, err := ctx.ReportGeneratedKey(nil, ctx.GetParticipantID(), generatedPubKey) // TODO: report the uncompressed?
		if err != nil {
			return t.failIfError(err, "failed to report GeneratedKey")
		}
		t.TxHash = txHash
		t.Status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckEthTx(*t.TxHash)
		ctx.GetLogger().Debug(fmt.Sprintf("id %v ReportGeneratedKey Status is %v", t.GetId(), status))
		if err != nil {
			return t.failIfError(err, "failed to check tx status")
		}
		if !core.IsPending(status) {
			t.Status = StatusDone
			return nil
		}
	}
	return nil
}

func (t *RequestAdded) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	msg = fmt.Sprintf("[%v] %v", t.GetId(), msg)
	return errors.Wrap(err, msg)
}
