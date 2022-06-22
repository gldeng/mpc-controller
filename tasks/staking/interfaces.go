package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/core"
)

type StakeTasker interface {
	TxHashGenerator
	SignatureSetter
	SignedTxGetter
}

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

type SignedTxGetter interface {
	GetSignedExportTx() (*evm.Tx, error)
	GetSignedImportTx() (*platformvm.Tx, error)
	GetSignedAddDelegatorTx() (*platformvm.Tx, error)
}

type Cache interface {
	cache.MyIndexGetter
	cache.GeneratedPubKeyInfoGetter
	cache.ParticipantKeysGetter
}

type Issuerer interface {
	IssueTask(ctx context.Context, task SignedTxGetter) ([]ids.ID, error)
}

type StakeTaskerCreatorer interface {
	CreateStakeTask() (StakeTasker, error)
}

type SignRequestCreatorer interface {
	CreateSignRequest(task TxHashGenerator) (*core.SignRequest, error)
}
