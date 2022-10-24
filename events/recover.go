package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/storage"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type OutputIndex uint32

const (
	OutputIndexPrincipal OutputIndex = iota
	OutputIndexReward
)

const (
	UTXOTypeRecover UTXOType = iota
)

type UTXOType int

type UTXOsFetched struct {
	NativeUTXOs []*avax.UTXO `json:"-"`
	MpcUTXOs    []*myAvax.MpcUTXO

	GroupIdHex       string         `copier:"must"`
	PartiIndex       *big.Int       `copier:"must"`
	GenPubKeyHex     string         `copier:"must"`
	GenPubKeyHashHex string         `copier:"must"` // key
	CChainAddress    common.Address `copier:"must"`
	PChainAddress    ids.ShortID    `copier:"must"`
}

type UTXOFetched struct {
	UTXO      *avax.UTXO
	GenPubKey storage.PubKey
}

type UTXOReported struct {
	NativeUTXO     *avax.UTXO `json:"-"`
	MpcUTXO        *myAvax.MpcUTXO
	TxHash         *common.Hash
	GenPubKeyBytes []byte
	GroupIDBytes   [32]byte
	PartiIndex     *big.Int
}

type ExportUTXORequest struct {
	TxID               ids.ID
	OutputIndex        uint32         `copier:"must"`
	To                 common.Address `copier:"must"`
	GenPubKeyHash      common.Hash
	ParticipantIndices []*big.Int `copier:"must"`
	TxHash             common.Hash
}

type UTXOHandled struct {
	NativeUTXO   *avax.UTXO `json:"-"`
	MpcUTXO      *myAvax.MpcUTXO
	ExportedTxID ids.ID
	ImportedTxID ids.ID
	Result       string
}
