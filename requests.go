// Requests around external services, which should be registered and issued with Manager.
// Request executors should reply with the given ReplyCh channel in a request,
// by doing so wo can make sure that each request will be responded respectively.

package mpc_controller

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Requests regarding contract interaction

type ReqToContractJoinRequest struct {
	Signer *bind.TransactOpts
	ReqId  *big.Int
	Index  *big.Int

	ReplyCh chan struct {
		Tx    *types.Transaction
		Error error
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Requests regarding C-Chain interaction

type ReqToCChainGetNonce struct {
	Account     common.Address
	BlockNumber *big.Int

	ReplyCh chan struct {
		Nonce uint64
		Error error
	}
}

type ReqToCChainGetBalance struct {
	Account     common.Address
	BlockNumber *big.Int

	ReplyCh chan struct {
		Balance *big.Int
		Error   error
	}
}

type ReqToCChainTransactionReceipt struct {
	TxHash common.Hash

	ReplyCh chan struct {
		Receipt *types.Receipt
		Error   error
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Requests regarding P-Chain interaction

type ReqToPChainIssueTx struct {
	ToChain string // enum: "C-Chain", "P-Chain"
	TxBytes []byte

	ReplyCh chan struct {
		ID    ids.ID
		Error error
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Requests regarding Mpc-Server interaction

type ReqToMpcServerReqKeygen struct {
	KeygenReq *core.KeygenRequest

	ReplyCh chan struct {
		Error error
	}
}

type ReqToMpcServerReqSign struct {
	SignReq *core.SignRequest

	ReplyCh chan struct {
		Error error
	}
}

type ReqToMpcServerGetResult struct {
	ReqId string // request id

	ReplyCh chan struct {
		Result *core.Result
		Error  error
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Requests regarding storage interaction

type ReqToStorerStoreGroupInfo struct {
	Group *storage.GroupInfo

	ReplyCh chan struct {
		Error error
	}
}

type ReqToStorerLoadGroupInfo struct {
	GroupIdHex string

	ReplyCh chan struct {
		*storage.GroupInfo
		Error error
	}
}

type ReqToStorerLoadGroupInfos struct {
	ReplyCh chan struct {
		Groups []*storage.GroupInfo
		Error  error
	}
}

type ReqToStorerStoreParticipantInfo struct {
	*storage.ParticipantInfo

	ReplyCh chan struct {
		Error error
	}
}

type ReqToStorerLoadParticipantInfo struct {
	PubKeyHashHex string
	GroupId       string

	ReplyCh chan struct {
		*storage.ParticipantInfo
		Error error
	}
}

type ReqToStorerLoadParticipantInfos struct {
	PubKeyHashHex string

	ReplyCh chan struct {
		Partis []*storage.ParticipantInfo
		Error  error
	}
}

type ReqToStorerGetParticipantIndex struct {
	PartiPubKeyHashHex string
	GenPubKeyHexHex    string

	ReplyCh chan struct {
		Index *big.Int
		Error error
	}
}

type ReqToStorerGetGroupIds struct {
	PartiPubKeyHashHex string

	ReplyCh chan struct {
		GroupIds [][32]byte
		Error    error
	}
}

type ReqToStorerGetPubKeys struct {
	PartiPubKeyHashHex string

	ReplyCh chan struct {
		PubKeys [][]byte
		Error   error
	}
}
