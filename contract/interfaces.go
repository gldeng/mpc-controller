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

type TransactorReportUTXO interface {
	ReportUTXO(ctx context.Context, groupId [32]byte, myIndex *big.Int, genPubKey []byte, utxoTxID [32]byte, utxoOutputIndex uint32) (*types.Transaction, error)
}

type TransactorJoinExportUTXO interface {
	JoinExportUTXO(ctx context.Context, groupId [32]byte, myIndex *big.Int, genPubKey []byte, utxoTxID [32]byte, utxoOutputIndex uint32) (*types.Transaction, error)
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

type FilterExportUTXORequestAdded interface {
	WatchExportUTXORequestAdded(ctx context.Context, publicKey [][]byte) (<-chan *MpcManagerExportUTXORequestAdded, error)
}

type FilterExportUTXORequestStarted interface {
	WatchExportUTXORequestStarted(ctx context.Context, publicKey [][]byte) (<-chan *MpcManagerExportUTXORequestStarted, error)
}
