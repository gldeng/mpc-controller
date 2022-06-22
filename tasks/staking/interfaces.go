package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/core"
)

type TxHasher interface {
	ExportTxHash() ([]byte, error)
	ImportTxHash() ([]byte, error)
	AddDelegatorTxHash() ([]byte, error)
}

type Cache interface {
	cache.MyIndexGetter
	cache.GeneratedPubKeyInfoGetter
	cache.ParticipantKeysGetter
}

type Issuerer interface {
	IssueTask(ctx context.Context, task *StakeTask) ([]ids.ID, error)
}

type StakeTaskCreatorer interface {
	CreateStakeTask() (*StakeTask, error)
}

type SignRequestCreatorer interface {
	CreateSignRequest(task TxHasher) (*core.SignRequest, error)
}
