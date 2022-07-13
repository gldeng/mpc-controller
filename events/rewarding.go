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

//type ExportUTXORequestAddedEvent struct {
//	TxID ids.ID
//	PublicKeyHash    common.Hash
//	TxHash           common.Hash
//}

//type JoinedExportUTXORequestEvent struct {
//	GroupIDHex       string
//	MyIndex          *big.Int
//	CompressedGenPubKeyHex        string
//	TxID ids.ID
//	TxHash           common.Hash
//}

type ExportUTXORequestEvent struct {
	TxID               ids.ID
	OutputIndex        uint32         `copier:"must"`
	To                 common.Address `copier:"must"`
	GenPubKeyHash      common.Hash
	ParticipantIndices []*big.Int `copier:"must"`
	TxHash             common.Hash
}

type UTXOExportedEvent struct {
	NativeUTXO   *avax.UTXO `json:"-"`
	MpcUTXO      *myAvax.MpcUTXO
	ExportedTxID ids.ID
	ImportedTxID ids.ID
}
