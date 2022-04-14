package task

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	avaEthclient "github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"strings"
	"time"
)

type MpcClient interface {
	Keygen(request *core.KeygenRequest) error
	Sign(request *core.SignRequest) error
	CheckResult(requestId string) (*core.Result, error)
}

type SignatureReceived struct {
	requestId string
	hash      string
	signature string
}

type PendingRequestId struct {
	taskId        string
	requestNumber uint8
}

//type SignRequest struct {
//	groupId [32]byte
//	publicKey string
//	participantKeys []string
//	hash string
//	// TODO: Add startTime to handle timeouts
//}

type SignResult struct {
	RequestId     string `json:"request_id"`
	Result        string `json:"result"`
	RequestType   string `json:"request_type"`
	RequestStatus string `json:"request_status"`
}

func parsePendingRequestId(str string) (*PendingRequestId, error) {
	var incorrectFormatErr = errors.New("PendingSignRequest is not in correct format")
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return nil, incorrectFormatErr
	}
	idIndex, requestNumberIndex := 0, 1
	var requestNumber uint8
	_, err := fmt.Sscan(parts[requestNumberIndex], &requestNumber)
	if err != nil {
		return nil, err
	}
	return &PendingRequestId{taskId: parts[idIndex], requestNumber: requestNumber}, nil
}

func (r *PendingRequestId) ToString() string {
	return fmt.Sprintf("%v/%v", r.taskId, r.requestNumber)
}

type TaskManager struct {
	networkContext        core.NetworkContext
	publicKeyCache        map[common.Hash]string
	myIndicesInGroups     map[string]*big.Int
	stakeTasks            map[string]*StakeTask
	pendingRequests       map[string]*core.SignRequest
	pendingKeygenRequests map[string]*core.KeygenRequest
	keygenRequestGroups   map[string][32]byte
	//networkID uint32
	//cchainID  ids.ID
	////assetID ids.ID
	//asset           avax.Asset
	avaEthclient    avaEthclient.Client
	myAddr          ids.ShortID
	coordinatorAddr common.Address
	wsClient        *ethclient.Client
	cChainClient    evm.Client
	eventsKA        chan *contract.MpcCoordinatorKeygenRequestAdded
	eventsStS       chan *contract.MpcCoordinatorStakeRequestStarted
	eventsStA       chan *contract.MpcCoordinatorStakeRequestAdded
	//mpcServiceUrl   string
	listener      *contract.MpcCoordinator
	instance      *contract.MpcCoordinator
	ethClient     *ethclient.Client
	secpFactory   avaCrypto.FactorySECP256K1R
	chSigReceived chan *SignatureReceived
	mpcClient     MpcClient
	transactor    *bind.TransactOpts
	myPubKey      string
	eventsPA      chan *contract.MpcCoordinatorParticipantAdded
	subPA         event.Subscription
	subKA         event.Subscription
	subStA        event.Subscription
	subStS        event.Subscription
	subKG         event.Subscription
	eventsKG      chan *contract.MpcCoordinatorKeyGenerated
}

func NewTaskManager(
	networkContext core.NetworkContext,
	mpcClient MpcClient,
	privateKey *ecdsa.PrivateKey,
	coordinatorAddr common.Address,
) (*TaskManager, error) {
	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, networkContext.ChainID())
	if err != nil {
		return nil, err
	}
	//pubKeyBytes := crypto.CompressPubkey(&privateKey.PublicKey)
	pubKeyBytes := marshalPubkey(&privateKey.PublicKey)[1:]
	pubKeyHex := common.Bytes2Hex(pubKeyBytes)
	hash := crypto.Keccak256Hash(pubKeyBytes)
	fmt.Printf("My PubKey is %v\n", pubKeyHex)
	fmt.Printf("My PubKey topic is %v\n", hash)
	return &TaskManager{
		networkContext:        networkContext,
		mpcClient:             mpcClient,
		transactor:            transactor,
		myPubKey:              pubKeyHex,
		coordinatorAddr:       coordinatorAddr,
		publicKeyCache:        make(map[common.Hash]string),
		myIndicesInGroups:     make(map[string]*big.Int),
		stakeTasks:            make(map[string]*StakeTask),
		pendingRequests:       make(map[string]*core.SignRequest),
		pendingKeygenRequests: make(map[string]*core.KeygenRequest),
		keygenRequestGroups:   make(map[string][32]byte),
	}, nil
}

func (m *TaskManager) Initialize() error {
	//cChainClient, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")
	cChainClient := evm.NewClient("http://localhost:9650", "C")
	wsClient, err := ethclient.Dial("ws://127.0.0.1:9650/ext/bc/C/ws")
	ethClient, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")
	if err != nil {
		return err
	}
	listener, err := contract.NewMpcCoordinator(m.coordinatorAddr, wsClient)
	if err != nil {
		return err
	}
	instance, err := contract.NewMpcCoordinator(m.coordinatorAddr, ethClient)
	if err != nil {
		return err
	}
	m.listener = listener
	m.instance = instance
	m.ethClient = ethClient
	m.cChainClient = cChainClient
	m.wsClient = wsClient
	m.chSigReceived = make(chan *SignatureReceived)
	m.eventsPA = make(chan *contract.MpcCoordinatorParticipantAdded)
	m.eventsKA = make(chan *contract.MpcCoordinatorKeygenRequestAdded)
	m.eventsKG = make(chan *contract.MpcCoordinatorKeyGenerated)
	m.eventsStA = make(chan *contract.MpcCoordinatorStakeRequestAdded)
	m.eventsStS = make(chan *contract.MpcCoordinatorStakeRequestStarted)
	m.secpFactory = avaCrypto.FactorySECP256K1R{}

	return nil
}

func (m *TaskManager) Start() error {
	err := m.subscribeParticipantAdded()
	if err != nil {
		return err
	}
	for {
		select {
		case evt, ok := <-m.eventsPA:
			if ok {
				fmt.Printf("ParticipantAdded: %v\n", evt)
				fmt.Printf("GroupId is %v\n", common.Bytes2Hex(evt.GroupId[:]))
				m.onParticipantAdded(evt)
			}
		case evt, ok := <-m.eventsKA:
			if ok {
				fmt.Printf("KeygenAdded %v\n", evt)
				err = m.onKeygenRequestAdded(evt)
				if err != nil {
					fmt.Errorf("failed to handle keygen %v", err)
				}
			} else {
				break
			}
		case evt, ok := <-m.eventsKG:
			if ok {
				fmt.Printf("KeyGenerated %v\n", evt)
			} else {
				break
			}
		case evt, ok := <-m.eventsStA:
			if ok {
				fmt.Printf("StakeRequestAdded %v\n", evt)
			} else {
				break
			}
		case evt, ok := <-m.eventsStS:
			if ok {
				fmt.Printf("StakeRequestStarted %v\n", evt)
			} else {
				break
			}
		case <-time.After(1 * time.Second):
			m.tick() // TODO: Handle error
		}
	}
}

func (m *TaskManager) tick() error {
	for requestId, _ := range m.pendingKeygenRequests {
		err := m.checkKeygenResult(requestId)
		if err != nil {
			return err
		}
	}
	for requestId, _ := range m.pendingRequests {
		err := m.checkResult(requestId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *TaskManager) checkKeygenResult(requestId string) error {
	result, err := m.mpcClient.CheckResult(requestId)
	if err != nil {
		return err
	}
	if result.RequestStatus == "DONE" {
		pubkey := common.Hex2Bytes(result.Result)
		groupId := m.keygenRequestGroups[requestId]
		ind, err := m.getMyIndexInGroup(groupId)
		if err != nil {
			return err
		}
		_, err = m.instance.ReportGeneratedKey(m.transactor, groupId, ind, pubkey)
		if err != nil {
			return err
		}
		delete(m.pendingKeygenRequests, requestId)
	}
	return nil
}

func (m *TaskManager) checkResult(requestId string) error {
	result, err := m.mpcClient.CheckResult(requestId)
	if err != nil {
		return err
	}
	if result.RequestStatus == "DONE" {
		var sig [65]byte
		sigBytes := common.Hex2Bytes(result.Result)
		copy(sig[:], sigBytes)
		reqId, err := parsePendingRequestId(requestId)
		if err != nil {
			return err
		}
		req := m.pendingRequests[requestId]
		task := m.stakeTasks[reqId.taskId]
		var hashMismatchErr = errors.New("hash doesn't match")
		var wrongRequestNumberErr = errors.New("wrong request number")
		if reqId.requestNumber == 0 {
			hash, err := task.ExportTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hash)
			if req.Hash != hashHex {
				return hashMismatchErr
			}
			err = task.SetExportTxSig(sig)
			if err != nil {
				return err
			}
			delete(m.pendingRequests, requestId)
			reqId.requestNumber += 1
			req.Hash = hashHex
			err = m.requestSign(reqId.ToString(), req)
			if err != nil {
				return err
			}
			m.pendingRequests[reqId.ToString()] = req
		} else if reqId.requestNumber == 1 {
			hash, err := task.ImportTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hash)
			if req.Hash != hashHex {
				return hashMismatchErr
			}
			err = task.SetImportTxSig(sig)
			if err != nil {
				return err
			}
			delete(m.pendingRequests, requestId)
			reqId.requestNumber += 1
			req.Hash = hashHex
			err = m.requestSign(reqId.ToString(), req)
			if err != nil {
				return err
			}
			m.pendingRequests[reqId.ToString()] = req
		} else if reqId.requestNumber == 2 {
			hash, err := task.AddDelegatorTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hash)
			if req.Hash != hashHex {
				return hashMismatchErr
			}
			err = task.SetAddDelegatorTxSig(sig)
			if err != nil {
				return err
			}
			delete(m.pendingRequests, requestId)
		} else {
			return wrongRequestNumberErr
		}

	}
	return nil
}

func (m *TaskManager) onKeygenRequestAdded(evt *contract.MpcCoordinatorKeygenRequestAdded) error {
	// Request MPC server
	group, err := m.instance.GetGroup(nil, evt.GroupId)
	if err != nil {
		return err
	}
	var participantKeys []string
	for _, k := range group.Participants {
		pk := common.Bytes2Hex(k)
		participantKeys = append(participantKeys, pk)
	}
	request := &core.KeygenRequest{
		RequestId:       evt.Raw.TxHash.Hex(),
		ParticipantKeys: participantKeys,

		Threshold: group.Threshold.Uint64(),
	}
	err = m.mpcClient.Keygen(request)
	if err != nil {
		return err
	}
	m.pendingKeygenRequests[request.RequestId] = request
	m.keygenRequestGroups[request.RequestId] = evt.GroupId
	return nil
}

func (m *TaskManager) subscribeParticipantAdded() error {
	if m.subPA != nil {
		m.subPA.Unsubscribe()
		m.subPA = nil
	}
	pubkey := common.Hex2Bytes(m.myPubKey)
	pubkeys := [][]byte{
		pubkey,
	}
	sub, err := m.listener.WatchParticipantAdded(nil, m.eventsPA, pubkeys)
	if err != nil {
		return err
	}
	m.subPA = sub
	return nil
}

func (m *TaskManager) subscribeKeygenRequestAdded() error {
	if m.subKA != nil {
		m.subKA.Unsubscribe()
		m.subKA = nil
	}
	var groupIds [][32]byte
	for groupIdHex, _ := range m.myIndicesInGroups {
		var groupId [32]byte
		groupIdRaw := common.Hex2BytesFixed(groupIdHex, 32)
		copy(groupId[:], groupIdRaw)
		groupIds = append(groupIds, groupId)
	}
	sub, err := m.listener.WatchKeygenRequestAdded(nil, m.eventsKA, groupIds)
	if err != nil {
		return err
	}
	m.subKA = sub
	return nil
}

func (m *TaskManager) subscribeKeyGenerated() error {
	if m.subKG != nil {
		m.subKG.Unsubscribe()
		m.subKG = nil
	}
	var groupIds [][32]byte
	for groupIdHex, _ := range m.myIndicesInGroups {
		var groupId [32]byte
		groupIdRaw := common.Hex2BytesFixed(groupIdHex, 32)
		copy(groupId[:], groupIdRaw)
		groupIds = append(groupIds, groupId)
	}
	sub, err := m.listener.WatchKeyGenerated(nil, m.eventsKG, groupIds)
	if err != nil {
		return err
	}
	m.subKG = sub
	return nil
}

func (m *TaskManager) subscribeStakeRequestAdded() error {
	if m.subStA != nil {
		m.subStA.Unsubscribe()
		m.subStA = nil
	}
	var pubkeys [][]byte
	for _, pubKeyHex := range m.publicKeyCache {
		pk := common.Hex2Bytes(pubKeyHex)
		pubkeys = append(pubkeys, pk)
	}
	sub, err := m.listener.WatchStakeRequestAdded(nil, m.eventsStA, pubkeys)
	if err != nil {
		return err
	}
	m.subStA = sub
	return nil
}

func (m *TaskManager) subscribeStakeRequestStarted() error {
	if m.subStS != nil {
		m.subStS.Unsubscribe()
		m.subStS = nil
	}
	var pubkeys [][]byte
	for _, pubKeyHex := range m.publicKeyCache {
		pk := common.Hex2Bytes(pubKeyHex)
		pubkeys = append(pubkeys, pk)
	}
	sub, err := m.listener.WatchStakeRequestStarted(nil, m.eventsStS, pubkeys)
	if err != nil {
		return err
	}
	m.subStS = sub
	return nil
}

func (m *TaskManager) onParticipantAdded(evt *contract.MpcCoordinatorParticipantAdded) {
	groupId := common.Bytes2Hex(evt.GroupId[:])
	m.myIndicesInGroups[groupId] = evt.Index
	m.subscribeKeygenRequestAdded()
	m.subscribeKeyGenerated()
}

func (m *TaskManager) onKeyGenerated(req *contract.MpcCoordinatorKeyGenerated) error {
	hash := crypto.Keccak256Hash(req.PublicKey)
	pkHex := common.Bytes2Hex(req.PublicKey)
	m.publicKeyCache[hash] = pkHex

	m.subscribeStakeRequestAdded()
	m.subscribeStakeRequestStarted()
	return nil
}

func (m *TaskManager) onStakeRequestAdded(req *contract.MpcCoordinatorStakeRequestAdded) error {
	pubKey := m.getPublicKey(req.PublicKey)
	ind, err := m.getMyIndex(pubKey)
	if err != nil {
		return err
	}
	_, err = m.instance.JoinRequest(m.transactor, req.RequestId, ind)
	if err != nil {
		return err
	}
	return nil
}

func (m *TaskManager) onStakeRequestStarted(req *contract.MpcCoordinatorStakeRequestStarted) error {
	// Request MPC server
	pubKey := m.getPublicKey(req.PublicKey)
	myInd, err := m.getMyIndex(pubKey)
	if err != nil {
		return err
	}

	var participating bool
	for _, ind := range req.ParticipantIndices {
		participating = participating || ind == myInd
	}
	if !participating {
		// Not Participating, Ignore
		return nil
	}

	nodeID, err := ids.ShortFromPrefixedString(req.NodeID, constants.NodeIDPrefix)

	if err != nil {
		return err
	}

	if pkHex, ok := m.publicKeyCache[req.PublicKey]; ok {
		pkBytes := common.Hex2Bytes(pkHex)

		pk, err := unmarshalPubkey(pkBytes)
		if err != nil {
			return err
		}
		cChainAddress := crypto.PubkeyToAddress(*pk)
		nonce, err := m.ethClient.NonceAt(context.Background(), cChainAddress, nil)

		if err != nil {
			return err
		}

		var invalidUint64Err = errors.New("invalid uint64")
		baseFeeGwei := uint64(300) // TODO: It should be given by the contract
		if !req.Amount.IsUint64() || !req.StartTime.IsUint64() || !req.EndTime.IsUint64() {
			return invalidUint64Err
		}
		task, err := NewStakeTask(m.networkContext, *pk, nonce, nodeID, req.Amount.Uint64(), req.StartTime.Uint64(), req.EndTime.Uint64(), baseFeeGwei)
		if err != nil {
			return err
		}
		taskId := req.Raw.TxHash.Hex()
		m.stakeTasks[taskId] = task
		hashBytes, err := task.ExportTxHash()
		if err != nil {
			return err
		}
		pariticipantKeys, err := m.getPariticipantKeys(pubKey, req.ParticipantIndices)
		if err != nil {
			return err
		}
		reqId := PendingRequestId{taskId: taskId, requestNumber: 0}
		hash := common.Bytes2Hex(hashBytes)
		request := &core.SignRequest{
			RequestId:       reqId.ToString(),
			PublicKey:       pubKey,
			ParticipantKeys: pariticipantKeys,
			Hash:            hash,
		}
		err = m.mpcClient.Sign(request)
		if err != nil {
			return err
		}
		m.pendingRequests[taskId] = request
	}
	return nil
}

func (m *TaskManager) requestKeygen(req *contract.MpcCoordinatorKeygenRequestAdded) error {
	/*
		m.mpcClient.Keygen(core.KeygenRequest{RequestId: req.Raw.TxHash.Hex(), })
		ParticipantKeys
		res, err := m.instance.GetGroup(nil, req.GroupId)
		if err != nil {
			return err
		}
		t := res.Threshold.String()
		id := req.Raw.TxHash.Hex()
		pubKeys := ""
		for i, pk := range res.Participants {
			var pref string
			if pref = ""; i > 0 {
				pref = ","
			}
			pubKeys += fmt.Sprintf(`%v"%v"`, pref, common.Bytes2Hex(pk))
		}
		payloadStr := fmt.Sprintf(`{"request_id": "%v", "public_keys": [%v], "t": %v}`, id, pubKeys, t)
		payload := strings.NewReader(payloadStr)
		http.Post(m.mpcServiceUrl+"/keygen", "application/json", payload)

	*/
	return nil
}

func (m *TaskManager) requestSign(requestId string, request *core.SignRequest) error {
	/*
		pubKeys := ""
		cnt := 0
		for _, k := range request.participantKeys {
			var pref string
			if pref = ""; cnt > 0 {
				pref = ","
			}
			pubKeys += fmt.Sprintf(`%v"%v"`, pref, k)
			cnt += 1
		}
		payloadStr := fmt.Sprintf(`{"request_id": "%v", "public_key": "%v", "hash": "%v", "participant_public_keys": [%v]}`, requestId, request.publicKey, request.hash, pubKeys)
		payload := strings.NewReader(payloadStr)
		http.Post(m.mpcServiceUrl+"/sign", "application/json", payload)

	*/
	return nil
}

func (m *TaskManager) getPublicKey(topic common.Hash) string {
	return m.publicKeyCache[topic]
}

func (m *TaskManager) getMyIndex(publicKey string) (*big.Int, error) {

	k := common.Hex2Bytes(publicKey)

	inf, err := m.instance.GetKey(nil, k)
	if err != nil {
		return nil, err
	}

	group, err := m.instance.GetGroup(nil, inf.GroupId)
	if err != nil {
		return nil, err
	}
	for i, pkBytes := range group.Participants {
		pk := common.Bytes2Hex(pkBytes)
		if publicKey == pk {
			return big.NewInt(int64(i) + 1), nil
		}
	}
	return nil, errors.New("not a member of the group")
}

func (m *TaskManager) getMyIndexInGroup(groupId [32]byte) (*big.Int, error) {

	group, err := m.instance.GetGroup(nil, groupId)
	if err != nil {
		return nil, err
	}
	for i, pkBytes := range group.Participants {
		pk := common.Bytes2Hex(pkBytes)
		if m.myPubKey == pk {
			return big.NewInt(int64(i) + 1), nil
		}
	}
	return nil, errors.New("not a member of the group")
}

func (m *TaskManager) getPariticipantKeys(publicKey string, indices []*big.Int) ([]string, error) {

	k := common.Hex2Bytes(publicKey)

	inf, err := m.instance.GetKey(nil, k)
	if err != nil {
		return nil, err
	}

	group, err := m.instance.GetGroup(nil, inf.GroupId)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, ind := range indices {
		k := group.Participants[ind.Uint64()]
		pk := common.Bytes2Hex(k)
		out = append(out, pk)
	}
	return out, nil
}

func unmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	var errInvalidPubkey = errors.New("invalid secp256k1 public key")
	if pub[0] == 4 {
		x, y := elliptic.Unmarshal(crypto.S256(), pub)
		if x == nil {
			return nil, errInvalidPubkey
		}
		return &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}, nil
	} else {
		x, y := elliptic.UnmarshalCompressed(crypto.S256(), pub)
		if x == nil {
			return nil, errInvalidPubkey
		}
		return &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}, nil
	}
}

func marshalPubkey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
}
