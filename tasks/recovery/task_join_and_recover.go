package recovery

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/tasks/join"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
	"time"
)

var (
	_ core.Task = (*JoinAndRecover)(nil)
)

type JoinAndRecover struct {
	Event            contract.MpcManagerRequestFailed
	RequestInitiated bool
	Request          Request
	ReqHash          types.RequestHash
	GroupId          [32]byte
	GroupInfoLoaded  bool
	Threshold        int64
	Indices          []uint
	PublicKey        []byte

	Join          *join.Join
	Recover       *Recovery
	StartTime     *time.Time
	QuorumReached bool
	Failed        bool
	Done          bool
	NoNeedToRun   bool
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
		PublicKey:        nil,
		Join:             nil,
		Recover:          nil,
		StartTime:        nil,
		QuorumReached:    false,
		Failed:           false,
		Done:             false,
		NoNeedToRun:      false,
	}

	return h, nil
}

func (t *JoinAndRecover) GetId() string {
	flowID := "join_and_recover" + "_" + bytes.BytesToHex(t.Event.RequestHash[:])
	return fmt.Sprintf("JoinAndRecover(%x)", flowID)
}

func (t *JoinAndRecover) IsDone() bool {
	return t.NoNeedToRun || (t.Recover != nil && t.Recover.IsDone())
}

func (t *JoinAndRecover) FailedPermanently() bool {
	if t.Recover != nil {
		return t.Recover.FailedPermanently()
	}
	return t.Failed
}

func (t *JoinAndRecover) IsSequential() bool {
	if t.Recover != nil {
		return t.Recover.IsSequential()
	}
	return true
}

func (t *JoinAndRecover) Next(ctx core.TaskContext) ([]core.Task, error) {
	if !t.RequestInitiated {
		txId, _ := ctx.GetTxIndex().GetTxByType(t.Request.OriginalRequestHash, core.TxTypeAddDelegator)
		if txId != ids.Empty {
			t.NoNeedToRun = true
			return nil, nil
		}
		pubkey, err := t.getPubKey(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to get public key")
		}
		t.PublicKey = pubkey
		err = t.createRequest(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to create request")
		}
		err = t.loadGroupInfo(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to load group info")
		}
		t.RequestInitiated = true
	}

	if t.Recover == nil {
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

		recoveryTask, err := NewRecovery(&t.Request, *quorumInfo)
		if err != nil {
			prom.FlowInitErr.With(prometheus.Labels{"flow": "recovery"}).Inc()
			return nil, t.failIfErrorf(err, "create recovery task")
		}
		t.Recover = recoveryTask
		prom.FlowInit.With(prometheus.Labels{"flow": "recovery"}).Inc()
		_, err = t.Recover.Next(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to run InitialStake")
		}
	} else {
		_, err := t.Recover.Next(ctx)
		if err != nil {
			return nil, t.failIfErrorf(err, "failed to run InitialStake")
		}
	}

	return nil, nil
}

func (t *JoinAndRecover) createRequest(ctx core.TaskContext) error {
	exportTxID := ids.ID{}
	copy(exportTxID[:], t.Event.Data[:])
	request := Request{
		OriginalRequestHash: t.Event.RequestHash,
		ExportTxID:          exportTxID,
	}
	t.Request = request

	hash, err := request.Hash()
	if err != nil {
		return errors.Wrap(err, "failed to get request hash")
	}
	t.ReqHash = hash
	return nil
}

func (t *JoinAndRecover) joinAndWaitUntilQuorumReached(ctx core.TaskContext) error {
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
				prom.TaskTimeout.With(prometheus.Labels{"flow": "", "task": "JoinAndRecover"}).Inc()
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

func (t *JoinAndRecover) getQuorumInfo(ctx core.TaskContext) (*types.QuorumInfo, error) {
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
		PubKey:             t.PublicKey,
	}, nil

}

func (t *JoinAndRecover) loadGroupInfo(ctx core.TaskContext) error {
	if !t.GroupInfoLoaded {
		groupId, err := ctx.GetGroupIdByKey(nil, t.PublicKey)
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

func (t *JoinAndRecover) getConfirmationCount(ctx core.TaskContext) (int64, error) {
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

func (t *JoinAndRecover) isJoinDone() bool {
	if t.Join == nil {
		return false
	}
	return t.Join.IsDone()
}

func (t *JoinAndRecover) isJoinFailed() bool {
	if t.Join == nil {
		return false
	}
	return t.Join.FailedPermanently()
}

func (t *JoinAndRecover) initJoin(ctx core.TaskContext) error {
	join := join.NewJoin(t.ReqHash)
	if join == nil {
		t.Failed = true
		return errors.Errorf("invalid sub joining task created for request %+x", t.ReqHash)
	}
	t.Join = join
	return nil
}

func (t *JoinAndRecover) saveRequest(ctx core.TaskContext) error {

	rBytes, err := t.Request.Encode()
	if err != nil {
		return err
	}

	key := []byte("request/")
	key = append(key, t.ReqHash[:]...)
	return ctx.GetDb().Set(context.Background(), key, rBytes)
}

func (t *JoinAndRecover) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *JoinAndRecover) getPubKey(ctx core.TaskContext) ([]byte, error) {
	txId, _ := ctx.GetTxIndex().GetTxByType(t.Request.OriginalRequestHash, core.TxTypeImportP)
	if txId != ids.Empty {
		tx, err := ctx.GetPChainTx(txId)
		if err != nil {
			return nil, t.failIfErrorf(err, ErrMsgFailedToRetrieveTx)
		}

		unsignedBytes := tx.Bytes()
		hash := hashing.ComputeHash256(unsignedBytes)
		cred := tx.Creds[0].(*secp256k1fx.Credential)
		pk, err := new(avaCrypto.FactorySECP256K1R).RecoverHashPublicKey(hash, cred.Sigs[0][:])
		return pk.Bytes(), nil
	}
	if t.Request.ExportTxID != ids.Empty {
		tx, err := ctx.GetCChainTx(t.Request.ExportTxID)
		if err != nil {
			return nil, t.failIfErrorf(err, ErrMsgFailedToRetrieveTx)
		}
		unsignedBytes := tx.Bytes()
		hash := hashing.ComputeHash256(unsignedBytes)
		cred := tx.Creds[0].(*secp256k1fx.Credential)
		pk, err := new(avaCrypto.FactorySECP256K1R).RecoverHashPublicKey(hash, cred.Sigs[0][:])
		return pk.Bytes(), nil
	}
	err := errors.New(ErrMsgFailedToDecidePubKey)
	return nil, t.failIfErrorf(err, "")
}

func extractThreshold(groupId [32]byte) int64 {
	n := new(big.Int)
	n.SetBytes(groupId[:])
	n.Rsh(n, 8)
	n.And(n, big.NewInt(255))
	return n.Int64()
}
