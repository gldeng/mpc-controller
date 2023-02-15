package stake

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
	"time"

	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/tasks/join"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*JoinAndStake)(nil)
)

type JoinAndStake struct {
	Event            contract.MpcManagerStakeRequestAdded
	RequestInitiated bool
	Request          Request
	ReqHash          types.RequestHash
	GroupId          [32]byte
	GroupInfoLoaded  bool
	Threshold        int64
	Indices          []uint

	Join          *join.Join
	InitialStake  *InitialStake
	StartTime     *time.Time
	QuorumReached bool
	Failed        bool
	Done          bool
}

func (t *JoinAndStake) GetId() string {
	id, _ := t.Request.Hash()
	flowID := "initialStake" + "_" + t.Event.RequestNumber.String() + "_" + id.String()
	return fmt.Sprintf("JoinAndStake(%x)", flowID)
}

func (t *JoinAndStake) IsDone() bool {
	return t.InitialStake != nil && t.InitialStake.IsDone()
}

func (t *JoinAndStake) FailedPermanently() bool {
	if t.InitialStake != nil {
		return t.InitialStake.FailedPermanently()
	}
	return t.Failed
}

func (t *JoinAndStake) IsSequential() bool {
	if t.InitialStake != nil {
		return t.InitialStake.IsSequential()
	}
	return true
}

func NewStakeJoinAndStake(event contract.MpcManagerStakeRequestAdded) (*JoinAndStake, error) {

	h := &JoinAndStake{
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

func (t *JoinAndStake) Next(ctx core.TaskContext) ([]core.Task, error) {
	if !t.RequestInitiated {
		err := t.createRequest(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to create request")
		}
		err = t.loadGroupInfo(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to load group info")
		}
		t.RequestInitiated = true
	}

	if t.InitialStake == nil {
		err := t.initJoin(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to init join")
		}

		err = t.joinAndWaitUntilQuorumReached(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to join")
		}
		quorumInfo, err := t.getQuorumInfo(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to get quorum info")
		}

		partiKeys, genKey, err := quorumInfo.CompressKeys()
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to compress keys")
		}
		ctx.GetLogger().Debug("got quorum info", []logger.Field{
			{"genPubKey", genKey},
			{"genCChainAddr", quorumInfo.CChainAddress()},
			{"genPChainAddr", quorumInfo.PChainAddress()},
			{"partiPubKeys", partiKeys}}...)

		initStake, err := NewInitialStake(&t.Request, *quorumInfo)
		if err != nil {
			prom.FlowInitErr.With(prometheus.Labels{"flow": "initialStake"}).Inc()
			return nil, t.failIfErrorf(err, "create InitialStake task")
		}
		t.InitialStake = initStake
		prom.FlowInit.With(prometheus.Labels{"flow": "initialStake"}).Inc()
		next, err := t.InitialStake.Next(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to run InitialStake")
		} else {
			return next, nil
		}
	} else {
		next, err := t.InitialStake.Next(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to run InitialStake")
		} else {
			return next, nil
		}
	}

	return nil, nil
}

func (t *JoinAndStake) createRequest(ctx core.TaskContext) error {
	pubKey, err := ctx.LastGenPubKey(nil)
	if err != nil {
		return err
	}
	request := Request{
		ReqNo:     t.Event.RequestNumber.Uint64(),
		TxHash:    common.Hash{},
		PubKey:    pubKey,
		NodeID:    t.Event.NodeID,
		Amount:    t.Event.Amount.String(),
		StartTime: t.Event.StartTime.Uint64(),
		EndTime:   t.Event.EndTime.Uint64(),
	}
	t.Request = request

	hash, err := request.Hash()
	if err != nil {
		return errors.Wrap(err, "failed to get request hash")
	}
	t.ReqHash = hash
	return nil
}

func (t *JoinAndStake) joinAndWaitUntilQuorumReached(ctx core.TaskContext) error {
	if t.StartTime == nil {
		now := time.Now()
		t.StartTime = &now
	}

	timeout := 60 * time.Minute
	interval := 2 * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop()
	var err error

	for {
		select {
		case <-timer.C:
			if time.Now().Sub(*t.StartTime) >= timeout {
				prom.TaskTimeout.With(prometheus.Labels{"flow": "", "task": "joinAndStake"}).Inc()
				return errors.New(ErrMsgTimedOut)
			}

			if t.isJoinFailed() {
				return errors.Wrap(err, "failed to join")
			}

			if !t.isJoinDone() {
				_, err := t.Join.Next(ctx)
				if err != nil || t.Join.FailedPermanently() {
					return t.failIfErrorf(err, "failed to run join")
				}
				if t.Join.IsDone() {
					prom.MpcJoinStake.Inc()
				}
			}

			if !t.QuorumReached {
				count, err := t.getConfirmationCount(ctx)
				if err != nil {
					return t.failIfErrorf(err, "failed to get confirmation count")
				}
				if count == t.Threshold+1 {
					prom.MpcJoinStakeQuorumReached.Inc()
					t.QuorumReached = true
					return nil // Done without error
				}
			}

			timer.Reset(interval)
		}
	}
}

func (t *JoinAndStake) getQuorumInfo(ctx core.TaskContext) (*types.QuorumInfo, error) {
	group, err := ctx.LoadGroup(t.GroupId)
	if err != nil {
		return nil, t.failIfErrorf(err, "failed to load group")
	}

	ctx.GetLogger().Debugf("loaded group id: %v %v", t.GroupId, group)
	ctx.GetLogger().Debugf("indices: %v", t.Indices)

	var pubKeys types.PubKeys
	for _, ind := range t.Indices {
		pubKeys = append(pubKeys, group.MemberPublicKeys[ind-1])
	}

	return &types.QuorumInfo{
		ParticipantPubKeys: pubKeys,
		PubKey:             t.Request.PubKey,
	}, nil

}

func (t *JoinAndStake) loadGroupInfo(ctx core.TaskContext) error {
	if !t.GroupInfoLoaded {
		groupId, err := ctx.GetGroupIdByKey(nil, t.Request.PubKey)
		if err != nil {
			return err
		}
		t.GroupId = groupId
		ctx.GetLogger().Debugf("retrieved group id: %x", groupId)
		t.Threshold = extractThreshold(groupId)
		t.GroupInfoLoaded = true
	}
	return nil
}

func (t *JoinAndStake) getConfirmationCount(ctx core.TaskContext) (int64, error) {
	confirmation, err := ctx.RequestRecords(nil, t.GroupId, t.ReqHash)
	if err != nil {
		return 0, err
	}
	confirmCount := new(big.Int)
	confirmCount.And(confirmation, big.NewInt(255))
	indices := types.Indices(*confirmation)
	t.Indices = indices.Indices()
	return confirmCount.Int64(), nil
}

func (t *JoinAndStake) isJoinDone() bool {
	if t.Join == nil {
		return false
	}
	return t.Join.IsDone()
}

func (t *JoinAndStake) isJoinFailed() bool {
	if t.Join == nil {
		return false
	}
	return t.Join.FailedPermanently()
}

func (t *JoinAndStake) initJoin(ctx core.TaskContext) error {
	join := join.NewJoin(t.ReqHash)
	if join == nil {
		t.Failed = true
		return errors.Errorf("invalid sub joining task created for request %+x", t.ReqHash)
	}
	t.Join = join
	return nil
}

func (t *JoinAndStake) saveRequest(ctx core.TaskContext) error {

	rBytes, err := t.Request.Encode()
	if err != nil {
		return err
	}

	key := []byte("request/")
	key = append(key, t.ReqHash[:]...)
	return ctx.GetDb().Set(context.Background(), key, rBytes)
}

func (t *JoinAndStake) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func extractThreshold(groupId [32]byte) int64 {
	n := new(big.Int)
	n.SetBytes(groupId[:])
	n.Rsh(n, 8)
	n.And(n, big.NewInt(255))
	return n.Int64()
}
