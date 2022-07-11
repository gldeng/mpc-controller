package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning rewarding

type StakingPeriodEndedEvent struct {
	AddDelegatorTxID ids.ID `copier:"must"`

	RequestID   uint64     `copier:"must"`
	DelegateAmt uint64     `copier:"must"`
	StartTime   uint64     `copier:"must"`
	EndTime     uint64     `copier:"must"`
	NodeID      ids.NodeID `copier:"must"`

	PubKeyHex     string      `copier:"must"`
	PChainAddress ids.ShortID `copier:"must"`
}

type UTXOsFetchedEvent struct {
	NativeUTXOs []*avax.UTXO `json:"-"`
	MpcUTXOs    []*myAvax.MpcUTXO

	GroupIdHex       string         `copier:"must"`
	PartiIndex       *big.Int       `copier:"must"`
	GenPubKeyHex     string         `copier:"must"`
	GenPubKeyHashHex string         `copier:"must"`
	CChainAddress    common.Address `copier:"must"`
	PChainAddress    ids.ShortID    `copier:"must"`
}

type UTXOReportedEvent struct {
	TxID           ids.ID
	OutputIndex    uint32
	TxHash         *common.Hash
	GenPubKeyBytes []byte
	GroupIDBytes   [32]byte
	PartiIndex     *big.Int
}

type ExportRewardRequestAddedEvent struct {
	AddDelegatorTxID ids.ID
	PublicKeyHash    common.Hash
	TxHash           common.Hash
}

type JoinedExportRewardRequestEvent struct {
	GroupIDHex       string
	MyIndex          *big.Int
	PubKeyHex        string
	AddDelegatorTxID ids.ID
	TxHash           common.Hash
}

type ExportRewardRequestStartedEvent struct {
	AddDelegatorTxID   ids.ID
	PublicKeyHash      common.Hash
	ParticipantIndices []*big.Int
	TxHash             common.Hash
}

type RewardExportedEvent struct {
	AddDelegatorTxID ids.ID
	ExportedTxID     ids.ID
	ImportedTxID     ids.ID
	// todo: more fields
}
