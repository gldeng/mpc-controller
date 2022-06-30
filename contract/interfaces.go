package contract

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding MpcManager

// Callers

type CallerGetGroup interface {
	GetGroup(ctx context.Context, groupId [32]byte) (Participants [][]byte, Threshold *big.Int, err error)
}

// Transactor

type TransactorJoinRequest interface {
	JoinRequest(ctx context.Context, requestId *big.Int, myIndex *big.Int) (*types.Transaction, error)
}

type TransactorReportGeneratedKey interface {
	ReportGeneratedKey(ctx context.Context, groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error)
}

type TransactorReportRewardUTXOs interface {
	ReportRewardUTXOs(ctx context.Context, addDelegatorTxID [32]byte, rewardUTXOIDs [][32]byte) (*types.Transaction, error)
}

// Filters

type FilterParticipantAdded interface {
	WatchParticipantAdded(ctx context.Context, publicKey [][]byte) (<-chan *MpcManagerParticipantAdded, error)
}

type FilterKeygenRequestAdded interface {
	WatchKeygenRequestAdded(ctx context.Context, groupId [][32]byte) (<-chan *MpcManagerKeygenRequestAdded, error)
}

type FilterKeyGenerated interface {
	WatchKeyGenerated(ctx context.Context, groupId [][32]byte) (<-chan *MpcManagerKeyGenerated, error)
}

type FilterStakeRequestAdded interface {
	WatchStakeRequestAdded(ctx context.Context, publicKey [][]byte) (<-chan *MpcManagerStakeRequestAdded, error)
}

type FilterStakeRequestStarted interface {
	WatchStakeRequestStarted(ctx context.Context, publicKey [][]byte) (<-chan *MpcManagerStakeRequestStarted, error)
}
