package joining

import (
	"context"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/pkg/errors"
	"sync"
)

// Subscribe event: *events.StakeRequestAdded

type StakeRequestAdded struct {
	BoundTransactor transactor.Transactor
	DB              storage.DB
	Logger          logger.Logger
	PartiPubKey     storage.PubKey

	stakeRequestAddedChan chan *events.StakeRequestAdded
	once                  sync.Once
}

func (eh *StakeRequestAdded) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.stakeRequestAddedChan = make(chan *events.StakeRequestAdded, 1024)
		go eh.joinRequest(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.StakeRequestAdded:
		select {
		case <-ctx.Done():
			return
		case eh.stakeRequestAddedChan <- evt:
		}
	}
}

func (eh *StakeRequestAdded) joinRequest(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-eh.stakeRequestAddedChan:
			genPubKey := &storage.GeneratedPublicKey{}
			key := genPubKey.KeyFromHash(evt.PublicKey)
			err := eh.DB.MGet(ctx, key, genPubKey)
			if err != nil {
				eh.Logger.ErrorOnError(err, "failed to load generated public key", []logger.Field{{"key", key}}...)
				break
			}

			participant := storage.Participant{
				PubKey:  hash256.FromBytes(eh.PartiPubKey),
				GroupId: genPubKey.GroupId,
			}

			err = eh.DB.LoadModel(ctx, &participant)
			if err != nil {
				eh.Logger.ErrorOnError(err, "failed to load participant", []logger.Field{{"key", participant.Key()}}...)
				break
			}

			partiId := participant.ParticipantId()
			txHash := evt.Raw.TxHash
			_, _, err = eh.BoundTransactor.JoinRequest(ctx, partiId, txHash)
			if err != nil {
				switch errors.Cause(err).(type) {
				case *transactor.ErrTypQuorumAlreadyReached:
					eh.Logger.DebugOnError(err, "Join stake request not accepted", []logger.Field{{"reqHash", txHash}}...)
				case *transactor.ErrTypAttemptToRejoin:
					eh.Logger.DebugOnError(err, "Join stake request not accepted", []logger.Field{{"reqHash", txHash}}...)
				default:
					eh.Logger.ErrorOnError(err, "Failed to join state request", []logger.Field{{"reqHash", txHash}}...)
				}
				break
			}

			stakeReq := storage.StakeRequest{
				ReqNo:              evt.RequestNumber.Uint64(),
				TxHash:             txHash,
				NodeID:             evt.NodeID,
				Amount:             evt.Amount.String(),
				StartTime:          evt.StartTime.Int64(),
				EndTime:            evt.EndTime.Int64(),
				GeneratedPublicKey: genPubKey,
			}

			reqHash := stakeReq.ReqHash()

			joinReq := &storage.JoinRequest{
				ReqHash: reqHash,
				PartiId: partiId,
				Args:    &stakeReq,
			}

			err = eh.DB.SaveModel(ctx, joinReq)
			eh.Logger.ErrorOnError(err, "Failed to save JoinRequest for stake", []logger.Field{
				{"joinReq", joinReq}}...)

			eh.Logger.WarnOnTrue(float64(len(eh.stakeRequestAddedChan)) > float64(cap(eh.stakeRequestAddedChan))*0.8, "Too many stake request PENDED to join",
				[]logger.Field{{"pendedStakeReqs", len(eh.stakeRequestAddedChan)}}...)
			eh.Logger.Debug("Joined stake request", []logger.Field{
				{"reqNo", evt.RequestNumber}, {"reqHash", txHash}}...)
		}
	}
}
