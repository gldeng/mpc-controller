package joining

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/work"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"sync"
	"time"
)

var (
	ErrCannotJoin = errors.New("Cannot join anymore")
)

// Accept event: *contract.MpcManagerStakeRequestAdded

// Emit event:

type StakeRequestAddedEventHandler struct {
	ContractAddr    common.Address
	Logger          logger.Logger
	MyIndexGetter   cache.MyIndexGetter
	MyPubKeyHashHex string
	Publisher       dispatcher.Publisher
	Receipter       chain.Receipter
	Signer          *bind.TransactOpts
	Transactor      bind.ContractTransactor
	evtObjChan      chan *dispatcher.EventObject
	once            sync.Once
	ws              *work.Workshop
}

func (eh *StakeRequestAddedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.evtObjChan = make(chan *dispatcher.EventObject, 1024)
		eh.ws = work.NewWorkshop(eh.Logger, "joinRequest", time.Minute*10, 10)
		go eh.joinRequest(ctx)
	})

	select {
	case <-ctx.Done():
		return
	case eh.evtObjChan <- evtObj:
	}
}

func (eh *StakeRequestAddedEventHandler) joinRequest(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-eh.evtObjChan:
			evt, ok := evtObj.Event.(*contract.MpcManagerStakeRequestAdded)
			if !ok {
				break
			}
			genPubKeyHashHex := evt.PublicKey.Hex()
			myIndex := eh.MyIndexGetter.GetMyIndex(eh.MyPubKeyHashHex, genPubKeyHashHex)
			if myIndex == nil {
				break
			}

			eh.ws.AddTask(ctx, &work.Task{
				Args: []interface{}{myIndex, evt},
				Ctx:  ctx,
				WorkFns: []work.WorkFn{func(ctx context.Context, args interface{}) {
					myIndex := args.([]interface{})[0].(*big.Int)
					evt := args.([]interface{})[1].(*contract.MpcManagerStakeRequestAdded)
					err := eh.doJoinRequest(ctx, myIndex, evt)
					if err != nil {
						if errors.Is(err, ErrCannotJoin) {
							eh.Logger.DebugOnError(err, "Join request unaccepted", []logger.Field{
								{"reqId", evt.RequestId}}...)
							return
						}
						eh.Logger.ErrorOnError(err, "Failed to join request", []logger.Field{
							{"reqId", evt.RequestId}}...)
						return
					}
				}},
			})
		}
	}
}

func (eh *StakeRequestAddedEventHandler) doJoinRequest(ctx context.Context, myIndex *big.Int, req *contract.MpcManagerStakeRequestAdded) (err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return errors.WithStack(err)
	}

	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		var err error
		_, err = transactor.JoinRequest(eh.Signer, req.RequestId, myIndex)
		if err != nil {
			if strings.Contains(err.Error(), "Cannot join anymore") {
				return false, errors.WithStack(ErrCannotJoin)
			}
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	return errors.WithStack(err)
}
