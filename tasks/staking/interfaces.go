package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/core"
)

type TxHashGenerator interface {
	ExportTxHash() ([]byte, error)
	ImportTxHash() ([]byte, error)
	AddDelegatorTxHash() ([]byte, error)
}

type SignatureSetter interface {
	SetExportTxSig(sig [sigLength]byte) error
	SetImportTxSig(sig [sigLength]byte) error
	SetAddDelegatorTxSig(sig [sigLength]byte) error
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
	CreateSignRequest(task TxHashGenerator) (*core.SignRequest, error)
}
