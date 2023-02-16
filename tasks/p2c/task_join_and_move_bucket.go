package p2c

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
	"time"

	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/tasks/join"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*JoinAndMoveBucket)(nil)
)

type JoinAndMoveBucket struct {
	Bucket           types.UtxoBucket
	RequestInitiated bool
	Request          MoveBucketRequest
	ReqHash          types.RequestHash
	GroupId          [32]byte
	GroupInfoLoaded  bool
	Threshold        int64
	Indices          []uint

	Join          *join.Join
	MoveBucket    *MoveBucket
	StartTime     *time.Time
	QuorumReached bool
	Failed        bool
	Done          bool
}

func (t *JoinAndMoveBucket) GetId() string {
	flowID := fmt.Sprintf("move_bucket_%v_%v_%v", t.Request.Bucket.UtxoType.String(), t.Request.Bucket.StartTimestamp, t.Request.Bucket.EndTimestamp)
	return fmt.Sprintf("JoinAndMoveBucket(%x)", flowID)
}

func (t *JoinAndMoveBucket) IsDone() bool {
	return t.MoveBucket != nil && t.MoveBucket.IsDone()
}

func (t *JoinAndMoveBucket) FailedPermanently() bool {
	if t.MoveBucket != nil {
		return t.MoveBucket.FailedPermanently()
	}
	return t.Failed
}

func (t *JoinAndMoveBucket) IsSequential() bool {
	if t.MoveBucket != nil {
		return t.MoveBucket.IsSequential()
	}
	return true
}

func NewJoinAndMoveBucket(bucket types.UtxoBucket) (*JoinAndMoveBucket, error) {

	h := &JoinAndMoveBucket{
		Bucket:           bucket,
		RequestInitiated: false,
		Request:          MoveBucketRequest{},
		ReqHash:          types.RequestHash{},
		GroupId:          [32]byte{},
		GroupInfoLoaded:  false,
		Threshold:        0,
		Indices:          nil,
		Join:             nil,
		MoveBucket:       nil,
		StartTime:        nil,
		QuorumReached:    false,
		Failed:           false,
		Done:             false,
	}

	return h, nil
}

func (t *JoinAndMoveBucket) Next(ctx core.TaskContext) ([]core.Task, error) {
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

	if t.MoveBucket == nil {
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

		moveBucket, err := NewMoveBucket(&t.Request, *quorumInfo)
		if err != nil {
			prom.FlowInitErr.With(prometheus.Labels{"flow": "moveBucket"}).Inc()
			return nil, t.failIfErrorf(err, "create MoveBucket task")
		}
		t.MoveBucket = moveBucket
		prom.FlowInit.With(prometheus.Labels{"flow": "moveBucket"}).Inc()
		next, err := t.MoveBucket.Next(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to run MoveBucket")
		} else {
			return next, nil
		}
	} else {
		next, err := t.MoveBucket.Next(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to run MoveBucket")
		} else {
			return next, nil
		}
	}

	return nil, nil
}

func (t *JoinAndMoveBucket) createRequest(ctx core.TaskContext) error {
	request := MoveBucketRequest{
		Bucket: t.Bucket,
	}
	t.Request = request

	hash, err := request.Hash()
	if err != nil {
		return errors.Wrap(err, "failed to get request hash")
	}
	t.ReqHash = hash
	return nil
}

func (t *JoinAndMoveBucket) joinAndWaitUntilQuorumReached(ctx core.TaskContext) error {
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
				prom.TaskTimeout.With(prometheus.Labels{"flow": "", "task": "JoinAndMoveBucket"}).Inc()
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
					prom.MpcJoinMoveBucket.Inc()
				}
			}

			if !t.QuorumReached {
				count, err := t.getConfirmationCount(ctx)
				if err != nil {
					return t.failIfErrorf(err, "failed to get confirmation count")
				}
				if count == t.Threshold+1 {
					prom.MpcJoinMoveBucketQuorumReached.Inc()
					t.QuorumReached = true
					return nil // Done without error
				}
			}

			timer.Reset(interval)
		}
	}
}

func (t *JoinAndMoveBucket) getQuorumInfo(ctx core.TaskContext) (*types.QuorumInfo, error) {
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
		PubKey:             t.Request.Bucket.PublicKey,
	}, nil

}

func (t *JoinAndMoveBucket) loadGroupInfo(ctx core.TaskContext) error {
	if !t.GroupInfoLoaded {
		groupId, err := ctx.GetGroupIdByKey(nil, t.Request.Bucket.PublicKey)
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

func (t *JoinAndMoveBucket) getConfirmationCount(ctx core.TaskContext) (int64, error) {
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

func (t *JoinAndMoveBucket) isJoinDone() bool {
	if t.Join == nil {
		return false
	}
	return t.Join.IsDone()
}

func (t *JoinAndMoveBucket) isJoinFailed() bool {
	if t.Join == nil {
		return false
	}
	return t.Join.FailedPermanently()
}

func (t *JoinAndMoveBucket) initJoin(ctx core.TaskContext) error {
	join := join.NewJoin(t.ReqHash)
	if join == nil {
		t.Failed = true
		return errors.Errorf("invalid sub joining task created for request %+x", t.ReqHash)
	}
	t.Join = join
	return nil
}

func (t *JoinAndMoveBucket) saveRequest(ctx core.TaskContext) error {

	rBytes, err := t.Request.Encode()
	if err != nil {
		return err
	}

	key := []byte("request/")
	key = append(key, t.ReqHash[:]...)
	return ctx.GetDb().Set(context.Background(), key, rBytes)
}

func (t *JoinAndMoveBucket) failIfErrorf(err error, format string, a ...any) error {
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
