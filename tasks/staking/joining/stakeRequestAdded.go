package joining

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/pkg/errors"
	"sync"
)

var (
	ErrCannotJoin = errors.New("Cannot join anymore")
)

// Subscribe event: *events.MpcManagerStakeRequestAdded

// Publish event:

type StakeRequestAdded struct {
	Logger logger.Logger
	PubKey []byte

	DB         storage.DB
	Transactor transactor.Transactor

	mpcManagerStakeRequestAddedChan chan *contract.MpcManagerStakeRequestAdded
	once                            sync.Once
	//ws                 *work.Workshop
}

func (eh *StakeRequestAdded) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.mpcManagerStakeRequestAddedChan = make(chan *contract.MpcManagerStakeRequestAdded, 1024)
		//eh.ws = work.NewWorkshop(eh.Logger, "joinRequest", time.Minute*10, 10)
		go eh.joinRequest(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestAdded:
		select {
		case <-ctx.Done():
			return
		case eh.mpcManagerStakeRequestAddedChan <- evt:
		}
	}
}

func (eh *StakeRequestAdded) joinRequest(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-eh.mpcManagerStakeRequestAddedChan:
			//eh.ws.AddTask(ctx, &work.Task{
			//	Args: []interface{}{myIndex, evt},
			//	Ctx:  ctx,
			//	WorkFns: []work.WorkFn{func(ctx context.Context, args interface{}) {
			//		myIndex := args.([]interface{})[0].(*big.Int)
			//		evt := args.([]interface{})[1].(*contract.MpcManagerStakeRequestAdded)

			genPubKey := &storage.GeneratedPublicKey{}
			key := genPubKey.KeyFromHash(evt.PublicKey)
			err := eh.DB.MGet(ctx, key, genPubKey)
			if err != nil {
				eh.Logger.ErrorOnError(err, "failed to load generated public key", []logger.Field{{"key", key}}...)
				break
			}

			participant := storage.Participant{
				PubKey:  hash256.FromBytes(eh.PubKey),
				GroupId: genPubKey.GroupId,
			}

			err = eh.DB.LoadModel(ctx, &participant)
			if err != nil {
				eh.Logger.ErrorOnError(err, "failed to load participant", []logger.Field{{"key", participant.Key()}}...)
				break
			}

			txHash := evt.Raw.TxHash
			_, _, err = eh.Transactor.JoinRequest(ctx, participant.ParticipantId(), txHash)
			if err != nil {
				switch errors.Cause(err).(type) { // todo: exploring more concrete error types
				case *transactor.ErrTypQuorumAlreadyReached:
					eh.Logger.DebugOnError(err, "Join stake request not accepted", []logger.Field{
						{"reqNo", evt.RequestNumber}, {"txHash", txHash}}...)
				case *transactor.ErrTypAttemptToRejoin:
					eh.Logger.DebugOnError(err, "Join stake request not accepted", []logger.Field{
						{"reqNo", evt.RequestNumber}, {"txHash", txHash}}...)
				default:
					eh.Logger.ErrorOnError(err, "Failed to join state request", []logger.Field{
						{"reqNo", evt.RequestNumber}, {"txHash", txHash}}...)
				}
				break
			}

			eh.Logger.WarnOnTrue(float64(len(eh.mpcManagerStakeRequestAddedChan)) > float64(cap(eh.mpcManagerStakeRequestAddedChan))*0.8, "Too may stake request PENDED to join",
				[]logger.Field{{"pendedStakeReqs", len(eh.mpcManagerStakeRequestAddedChan)}}...)
			eh.Logger.Debug("Joined stake request", []logger.Field{
				{"reqNo", evt.RequestNumber}, {"txHash", txHash}}...)
			//}},
			//})
		}
	}
}
