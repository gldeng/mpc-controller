package recovery

import (
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/tasks/join"
	"time"
)

var (
	_ core.Task = (*JoinAndRecover)(nil)
)

type JoinAndRecover struct {
	Event            contract.MpcManagerRequestFailed
	RequestInitiated bool
	Request          types.RecoverRequest
	ReqHash          types.RequestHash
	GroupId          [32]byte
	GroupInfoLoaded  bool
	Threshold        int64
	Indices          []uint

	Join          *join.Join
	Recover       *Recovery
	StartTime     *time.Time
	QuorumReached bool
	Failed        bool
	Done          bool
}

func (t *JoinAndRecover) GetId() string {
	id, _ := t.Request.Hash()
	flowID := "initialStake" + "_" + t.Event.RequestHash.String() + "_" + id.String()
	return fmt.Sprintf("JoinAndStake(%x)", flowID)
}

func (t *JoinAndRecover) Next(ctx core.TaskContext) ([]core.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (t *JoinAndRecover) IsDone() bool {
	//TODO implement me
	panic("implement me")
}

func (t *JoinAndRecover) FailedPermanently() bool {
	//TODO implement me
	panic("implement me")
}

func (t *JoinAndRecover) IsSequential() bool {
	//TODO implement me
	panic("implement me")
}

func NewJoinAndRecover(event contract.MpcManagerRequestFailed) (*JoinAndRecover, error) {

	h := &JoinAndRecover{
		Event:            event,
		RequestInitiated: false,
		Request:          Request{},
		ReqHash:          types.RequestHash{},
		GroupId:          [32]byte{},
		GroupInfoLoaded:  false,
		Threshold:        0,
		Indices:          nil,
		Join:             nil,
		InitialStake:     nil,
		QuorumReached:    false,
		Failed:           false,
		Done:             false,
	}

	return h, nil
}
