package keygen

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
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
		ctx.GetLogger().Error(ErrMsgFailedToLoadGroup)
		return nil, t.failIfError(err, ErrMsgFailedToLoadGroup)
	}

	ctx.GetLogger().Debug(fmt.Sprintf("Loaded group %x", group.GroupId))
	t.group = group

	// TODO: improve retry strategy, involving further error handling
	//	interval := 100 * time.Millisecond
	//	timer := time.NewTimer(interval)
	//	defer timer.Stop() // TODO: stop all other timers to avoid resource leak
	//
	//loop:
	//	for {
	//		select {
	//		case <-timer.C:
	//			err = t.run(ctx)
	//			if t.Failed || t.Status == StatusDone {
	//				break loop
	//			}
	//
	//			timer.Reset(interval)
	//		}
	//	}
	//
	//	return nil, err
	return nil, t.run(ctx)
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
		var pubKeys types.PubKeys
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
			Threshold:              t.group.ParticipantID().Threshold(),
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
			return t.failIfError(err, "failed to get keygen result")
		}

		if res == nil || res.Result == "" {
			return t.failIfError(err, "empty keygen result")
		}

		if res.Status != core.StatusDone {
			ctx.GetLogger().Debug("Keygen not done")
			return nil
		}

		genPubKeyHex := res.Result
		decompressedPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(genPubKeyHex) // for Ethereum compatibility
		if err != nil {
			return t.failIfError(err, fmt.Sprintf("failed to decompress generated public key %v", genPubKeyHex))
		}

		txHash, err := ctx.ReportGeneratedKey(ctx.GetMyTransactSigner(), t.group.ParticipantID(), decompressedPubKeyBytes)
		if err != nil {
			return t.failIfError(err, "failed to report GeneratedKey")
		}
		t.TxHash = txHash
		t.Status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckEthTx(*t.TxHash)
		ctx.GetLogger().Debugf("id %v ReportGeneratedKey Status is %v", t.GetId(), status)
		if err != nil {
			ctx.GetLogger().Errorf("Failed to check status for tx %x : %v", *t.TxHash, err)
			return t.failIfError(err, fmt.Sprintf("failed to check status for tx %x", *t.TxHash))
		}

		switch status {
		case core.TxStatusUnknown:
			ctx.GetLogger().Debug(fmt.Sprintf("Unkonw tx status (%v:%x) of reporting generated key for group %x",
				status, *t.TxHash, t.group.GroupId))
			return t.failIfError(errors.Errorf("unkonw tx status (%v:%x) of reporting generated key for group %x",
				status, *t.TxHash, t.group.GroupId), "")
		case core.TxStatusAborted:
			t.Status = StatusKeygenReqSent // TODO: avoid endless repeating ReportGenerateKey?
			errMsg := fmt.Sprintf("ReportGeneratedKey tx %x aborted for group %x", *t.TxHash, t.group.GroupId)
			ctx.GetLogger().Error(errMsg)
			return errors.Errorf(errMsg)
		case core.TxStatusCommitted:
			t.Status = StatusDone
			ctx.GetLogger().Debug(fmt.Sprintf("Generated key reported for group %x", t.group.GroupId))
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
