package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning interact with contract

// MpcManager transactor

type ReportedGenPubKey struct { //    function reportGeneratedKey(bytes32 participantId, bytes calldata generatedPublicKey)
	GroupIdHex       string
	MyPartiIndex     *big.Int
	GenPubKeyHex     string
	GenPubKeyHashHex string
	CChainAddress    common.Address
	PChainAddress    ids.ShortID
}

type JoinRequest struct {
	RequestId  *big.Int
	PartiIndex *big.Int
}

type JoinedRequest struct { //    function joinRequest(bytes32 participantId, bytes32 requestHash) external onlyGroupMember(participantId) {
	TxHashHex  string
	RequestId  *big.Int
	PartiIndex *big.Int
}

// MpcManager emitted events

type ParticipantAdded struct { // event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
	PublicKey common.Hash // indexed
	GroupId   [32]byte
	Index     *big.Int
	Raw       types.Log
}

type KeygenRequestAdded struct { // event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
	GroupId [32]byte // indexed
	Raw     types.Log
}

type KeyGenerated struct { // event KeygenRequestAdded(bytes32 indexed groupId)
	GroupId   [32]byte // indexed
	PublicKey []byte
	Raw       types.Log
}

type StakeRequestAdded struct { // event StakeRequestAdded(uint256 requestNumber, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
	RequestNumber *big.Int
	PublicKey     common.Hash // indexed
	NodeID        string
	Amount        *big.Int
	StartTime     *big.Int
	EndTime       *big.Int
	Raw           types.Log
}

type RequestStarted struct { // event RequestStarted(bytes32 requestHash, uint256 participantIndices)
	ReqHash                *storage.RequestHash
	TaskType               storage.TaskType
	PartiIndices           *storage.Indices
	JoinedReq              *storage.JoinRequest
	CompressedPartiPubKeys []string
	CompressedGenPubKeyHex string
	Raw                    types.Log
}
