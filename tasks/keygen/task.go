package keygen

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	types2 "github.com/avalido/mpc-controller/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	Failed        bool
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
	if len(t.Event.Raw.Topics) < 2 {
		// Do nothing, invalid event
		return nil, nil
	}
	groupId := ctx.GetParticipantID().GroupId()
	if t.Event.Raw.Topics[1] != common.BytesToHash(crypto.Keccak256(groupId[:])) {
		// Not for me
		return nil, nil
	}

	interval := 100 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		select {
		case <-timer.C:
			err := t.run(ctx)
			if err != nil || t.Status == StatusDone {
				return nil, err
			} else {
				timer.Reset(interval)
			}
		}
	}

	return nil, nil
}

func (t *RequestAdded) IsDone() bool {
	return t.Status == StatusDone
}

func (t *RequestAdded) FailedPermanently() bool {
	//TODO implement me
	panic("implement me")
}

func (t *RequestAdded) RequiresNonce() bool {
	return false
}

func (t *RequestAdded) run(ctx core.TaskContext) error {
	switch t.Status {
	case StatusInit:
		key := []byte("group/")
		pId := ctx.GetParticipantID()
		groupId := pId.GroupId()
		key = append(key, groupId[:]...)
		groupBytes, err := ctx.GetDb().Get(context.Background(), key)
		if err != nil {
			return t.failIfError(err, "failed to get group")
		}
		group := &types2.Group{}
		err = group.Decode(groupBytes)
		if err != nil {
			return t.failIfError(err, "failed to decode group")
		}
		pubkeys := make([]string, 0)
		for _, publicKey := range group.MemberPublicKeys {
			pubkeys = append(pubkeys, common.Bytes2Hex(publicKey))
		}
		t.KeygenRequest = &core.KeygenRequest{
			ReqID:                  t.GetId(),
			CompressedPartiPubKeys: pubkeys,
			Threshold:              0,
		}
		err = ctx.GetMpcClient().Keygen(context.Background(), t.KeygenRequest)
		if err != nil {
			return t.failIfError(err, "failed to send keygen request")
		}
		t.Status = StatusKeygenReqSent
		return nil
	case StatusKeygenReqSent:
		res, err := ctx.GetMpcClient().Result(context.Background(), t.KeygenRequest.ReqID)
		// TODO: Handle 404
		if err != nil {
			return t.failIfError(err, "failed to get result")
		}

		if res.Status != core.StatusDone {
			ctx.GetLogger().Debug("keygen not done")
			return nil
		}
		generatedPubKey := common.Hex2Bytes(res.Result)
		txHash, err := ctx.ReportGeneratedKey(nil, ctx.GetParticipantID(), generatedPubKey)
		if err != nil {
			return t.failIfError(err, "failed to report GeneratedKey")
		}
		t.TxHash = txHash
		t.Status = StatusTxSent
		return nil
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
		return nil
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
