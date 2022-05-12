package task

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"

	//"errors"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	avaEthclient "github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

const sep = "-"

type ReportKeyTx struct {
	groupId            [32]byte
	myIndex            *big.Int
	generatedPublicKey []byte
}

type JoinTx struct {
	requestId *big.Int
	myIndex   *big.Int
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
	parts := strings.Split(str, sep)
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
	return fmt.Sprintf("%v"+sep+"%v", r.taskId, r.requestNumber)
}

type TaskManager struct {
	config config.Config
	log    logger.Logger
	storer storage.Storer
	staker *Staker

	mpcControllerId string
	networkContext  core.NetworkContext

	//groupCache map[string]*storage.GroupInfo

	//publicKeyCache    map[common.Hash]string
	//myIndicesInGroups map[string]*big.Int

	stakeTasks map[string]*StakeTask

	pendingSignRequests   map[string]*core.SignRequest
	pendingKeygenRequests map[string]*core.KeygenRequest

	//keygenRequestGroups map[string][32]byte

	pendingReports map[common.Hash]*ReportKeyTx

	pendingJoins map[common.Hash]*JoinTx

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
	mpcClient     core.MPCClient
	signer        *bind.TransactOpts
	myPubKey      string
	myPubKeyHash  common.Hash
	eventsPA      chan *contract.MpcCoordinatorParticipantAdded
	subPA         event.Subscription
	subKA         event.Subscription
	subStA        event.Subscription
	subStS        event.Subscription
	subKG         event.Subscription
	eventsKG      chan *contract.MpcCoordinatorKeyGenerated
}

func NewTaskManager(log logger.Logger, config config.Config, storer storage.Storer, staker *Staker,

//mpcControllerId int,
//networkContext core.NetworkContext,
//mpcClient core.MPCClient,
//privateKey *ecdsa.PrivateKey,
//coordinatorAddr common.Address,
) (*TaskManager, error) {
	//transactor, err := bind.NewKeyedTransactorWithChainID(config.ControllerKey_(), networkContext.ChainID())
	//if err != nil {
	//	return nil, _errors.Annotatef(err, "failed to create transaction signer")
	//}
	//pubKeyBytes := crypto.CompressPubkey(&privateKey.PublicKey)
	//log.With(logger.Field{"receiver", "task-manager"})

	privKey := config.ControllerKey()
	pubKeyBytes := marshalPubkey(&privKey.PublicKey)[1:]
	pubKeyHex := common.Bytes2Hex(pubKeyBytes)
	pubKeyHash := crypto.Keccak256Hash(pubKeyBytes)
	log.Debug("parsed task manager key info",
		logger.Field{"mpcControllerId", config.ControllerId()},
		logger.Field{"pubKey", pubKeyHex},
		logger.Field{"pubKeyTopic", pubKeyHash})
	m := &TaskManager{
		config:          config,
		log:             log,
		staker:          staker,
		storer:          storer,
		networkContext:  *config.NetworkContext(),
		mpcClient:       config.MpcClient(),
		signer:          config.ControllerSigner(),
		myPubKey:        pubKeyHex,
		myPubKeyHash:    pubKeyHash,
		coordinatorAddr: *config.CoordinatorAddress(),
		//groupCache:            make(map[string]*storage.GroupInfo),
		//publicKeyCache:        make(map[common.Hash]string),
		//myIndicesInGroups:     make(map[string]*big.Int),
		stakeTasks:            make(map[string]*StakeTask),
		pendingSignRequests:   make(map[string]*core.SignRequest),
		pendingKeygenRequests: make(map[string]*core.KeygenRequest),
		//keygenRequestGroups:   make(map[string][32]byte),
		pendingReports: make(map[common.Hash]*ReportKeyTx),
		pendingJoins:   make(map[common.Hash]*JoinTx),
	}

	m.listener = config.CoordinatorBoundListener()
	m.instance = config.CoordinatorBoundInstance()
	m.ethClient = config.EthRpcClient()
	//m.cChainClient = config.CChainIssueClient()
	//m.wsClient = config.EthWsClient()
	m.chSigReceived = make(chan *SignatureReceived)
	m.eventsPA = make(chan *contract.MpcCoordinatorParticipantAdded)
	m.eventsKA = make(chan *contract.MpcCoordinatorKeygenRequestAdded)
	m.eventsKG = make(chan *contract.MpcCoordinatorKeyGenerated)
	m.eventsStA = make(chan *contract.MpcCoordinatorStakeRequestAdded)
	m.eventsStS = make(chan *contract.MpcCoordinatorStakeRequestStarted)
	m.secpFactory = avaCrypto.FactorySECP256K1R{}
	return m, nil
}

////// todo: enable urls customizable
//func (m *TaskManager) Initialize() error {
//	//cChainClient, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")
//	//cChainClient := evm.NewClient("http://localhost:9650", "C")
//	//wsClient, err := ethclient.Dial("ws://127.0.0.1:9650/ext/bc/C/ws")
//	//if err != nil {
//	//	logger.Error("failed to dail ws://127.0.0.1:9650/ext/bc/C/ws", logger.Field{"ERROR", err})
//	//	os.Exit(1)
//	//}
//	ethClient, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")
//	if err != nil {
//		logger.Error("failed to dail http://localhost:9650/ext/bc/C/rpc", logger.Field{"ERROR", err})
//		os.Exit(1)
//	}
//	//listener, err := contract.NewMpcCoordinator(m.coordinatorAddr, wsClient)
//	//if err != nil {
//	//	logger.Error("failed to create ws mpc coordinator, ERROR: %v", logger.Field{"ERROR", err})
//	//	os.Exit(1)
//	//}
//	instance, err := contract.NewMpcCoordinator(m.coordinatorAddr, ethClient)
//	if err != nil {
//		logger.Error("failed to create mpc coordinator, ERROR: %v", logger.Field{"ERROR", err})
//		os.Exit(1)
//	}
//	//m.listener = listener
//	m.listener = m.config.CoordinatorBoundListener()
//
//	m.instance = instance
//	m.ethClient = ethClient
//	//m.cChainClient = cChainClient
//
//	//m.wsClient = wsClient
//	m.chSigReceived = make(chan *SignatureReceived)
//	m.eventsPA = make(chan *contract.MpcCoordinatorParticipantAdded)
//	m.eventsKA = make(chan *contract.MpcCoordinatorKeygenRequestAdded)
//	m.eventsKG = make(chan *contract.MpcCoordinatorKeyGenerated)
//	m.eventsStA = make(chan *contract.MpcCoordinatorStakeRequestAdded)
//	m.eventsStS = make(chan *contract.MpcCoordinatorStakeRequestStarted)
//	m.secpFactory = avaCrypto.FactorySECP256K1R{}
//
//	return nil
//}

// todo: logic to quit for loop

func (m *TaskManager) Start() error {
	err := m.subscribeParticipantAdded()
	if err != nil {
		return errors.WithStack(err)
	}
	for {
		select {
		case evt, ok := <-m.eventsPA:
			if !ok {
				m.log.Debug("Retrieve nothing from ParticipantAdded event channel")
				break
			}

			m.log.Info("Received ParticipantAdded event",
				logger.Field{"groupIdHex", common.Bytes2Hex(evt.GroupId[:])},
				logger.Field{"event", evt})

			err := m.onParticipantAdded(evt)
			if err != nil {
				m.log.Error("Failed to deal with event ParticipantAdded", logger.Field{"error", err})
			}

		case evt, ok := <-m.eventsKA:
			if !ok {
				m.log.Debug("Retrieve nothing from KeygenAdded event channel")
				break
			}

			m.log.Info("Received KeygenAdded event",
				logger.Field{"groupIdHex", common.Bytes2Hex(evt.GroupId[:])},
				logger.Field{"event", evt})

			err = m.onKeygenRequestAdded(evt)
			if err != nil {
				m.log.Error("Failed to respond to KeygenAdded event",
					logger.Field{"event", evt},
					logger.Field{"error", err})
			}

		case evt, ok := <-m.eventsKG:
			if !ok {
				m.log.Debug("Retrieve nothing from KeyGenerated event channel")
				break
			}

			m.log.Info("Received KeyGenerated event",
				logger.Field{"groupIdHex", common.Bytes2Hex(evt.GroupId[:])},
				logger.Field{"event", evt})

			err := m.onKeyGenerated(evt)
			if err != nil {
				m.log.Error("Failed to respond to KeyGenerated event",
					logger.Field{"event", evt},
					logger.Field{"error", err})
			}

		case evt, ok := <-m.eventsStA:
			if !ok {
				m.log.Debug("Retrieve nothing from StakeRequestAdded event channel")
				break
			}

			m.log.Info("Received StakeRequestAdded event",
				logger.Field{"event", evt})

			err := m.onStakeRequestAdded(evt)
			if err != nil {
				m.log.Error("Failed to respond to StakeRequestAdded event",
					logger.Field{"event", evt},
					logger.Field{"error", err})
			}

		case evt, ok := <-m.eventsStS:
			if !ok {
				m.log.Debug("Retrieve nothing from StakeRequestStarted event channel")
				break
			}

			m.log.Info("Received StakeRequestStarted event",
				logger.Field{"event", evt})

			//// Wait until the corresponding key has been generated
			//<-time.After(time.Second * 20)

			err := m.onStakeRequestStarted(evt)
			if err != nil {
				m.log.Error("Failed to respond to StakeRequestStarted event",
					logger.Field{"error", err})
			}

		case <-time.After(1 * time.Second):
			err := m.tick()
			if err != nil {
				m.log.Error("Got an tick error",
					logger.Field{"error", err})
				fmt.Printf("%+v", err)
			}
		}
	}
}

func (m *TaskManager) tick() error {
	err := m.checkPendingReports()
	if err != nil {
		return err
	}
	err = m.checkPendingJoins()
	if err != nil {
		return err
	}
	for requestId, _ := range m.pendingKeygenRequests {
		err := m.checkKeygenResult(requestId)
		if err != nil {
			return err
		}
	}
	for requestId, _ := range m.pendingSignRequests {
		err := m.checkSignResult(requestId)
		if err != nil {
			return err
		}
	}
	return nil
}

func sample(arr []common.Hash) []common.Hash {
	var out []common.Hash
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	for _, txHash := range arr {
		if r.Intn(1) == 0 {
			out = append(out, txHash)
		}
	}
	return out
}

func (m *TaskManager) checkPendingReports() error {

	var done []common.Hash
	var retry []common.Hash
	for txHash, _ := range m.pendingReports {
		rcp, err := m.ethClient.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			if rcp.Status == 1 {
				done = append(done, txHash)
			} else {
				retry = append(retry, txHash)
			}
		}
	}
	// TODO: Figure out why tx fails
	// Suspect due to contention between different users, for now make retry random
	sampledRetry := sample(retry)

	for _, txHash := range sampledRetry {
		req := m.pendingReports[txHash]
		groupId, ind, pubkey := req.groupId, req.myIndex, req.generatedPublicKey
		tx, err := m.instance.ReportGeneratedKey(m.signer, groupId, ind, pubkey)
		// todo: deal with error: "error": "insufficient funds for gas * price + value: address 0x3051bA2d313840932B7091D2e8684672496E9A4B have (2972700000107200) want (5436550000000000)
		if err != nil {
			m.log.Error("Failed to reported generated key",
				logger.Field{"error", err},
				logger.Field{"groupId", groupId},
				logger.Field{"pubKey", string(pubkey)})
			return errors.Wrapf(err, "failed to reported generated key %v for group %v", string(pubkey), groupId)
		}
		m.pendingReports[tx.Hash()] = &ReportKeyTx{
			groupId:            groupId,
			myIndex:            ind,
			generatedPublicKey: pubkey,
		}
		if err != nil {
			return err
		}
		fmt.Printf(
			"Retry reporting key tx:\n  groupId: %v\n  myIndex: %v\n  pubKey: %v\n  txHash: %v\n",
			common.Bytes2Hex(groupId[:]),
			ind,
			common.Bytes2Hex(pubkey),
			tx.Hash(),
		)
	}
	for _, txHash := range sampledRetry {
		delete(m.pendingReports, txHash)
	}
	for _, txHash := range done {
		delete(m.pendingReports, txHash)
	}
	return nil
}

func (m *TaskManager) checkPendingJoins() error {

	var done []common.Hash
	var retry []common.Hash
	for txHash, _ := range m.pendingJoins {
		rcp, err := m.ethClient.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			if rcp.Status == 1 {
				done = append(done, txHash)
			} else {
				retry = append(retry, txHash)
			}
		}
	}
	// TODO: Figure out why tx fails
	// Suspect due to contention between different users, for now make retry random
	sampledRetry := sample(retry)

	for _, txHash := range sampledRetry {
		req := m.pendingJoins[txHash]
		requestId, myIndex := req.requestId, req.myIndex
		tx, err := m.instance.JoinRequest(m.signer, requestId, myIndex)
		m.pendingJoins[tx.Hash()] = &JoinTx{
			requestId: requestId,
			myIndex:   myIndex,
		}
		if err != nil {
			return err
		}
		fmt.Printf(
			"Retry join tx:\n  requestId: %v\n  myIndex: %v\n  txHash: %v\n",
			requestId,
			myIndex,
			tx.Hash(),
		)
	}
	for _, txHash := range sampledRetry {
		delete(m.pendingJoins, txHash)
	}
	for _, txHash := range done {
		delete(m.pendingJoins, txHash)
	}
	return nil
}

func (m *TaskManager) checkKeygenResult(requestId string) error {
	result, err := m.mpcClient.Result(context.Background(), requestId) // todo: add shared context to task manager
	m.log.Debug("Task-manager retrieved keygen result from mpc-server",
		logger.Field{"reqId", requestId},
		logger.Field{"result", result})
	if err != nil {
		m.log.Error("Got an error when check key generating result",
			logger.Field{"requestId", requestId},
			logger.Field{"error", err})
		return err
	}
	if result.RequestStatus == "DONE" {
		genPubKey := common.Hex2Bytes(result.Result)

		keyGenInfo, err := m.storer.LoadKeygenRequestInfo(requestId)
		if err != nil {
			return errors.WithStack(err)
		}
		partyInfo, err := m.storer.LoadParticipantInfo(m.myPubKeyHash.Hex(), keyGenInfo.GroupIdHex)
		if err != nil {
			return errors.WithStack(err)
		}

		groupIdRaw := common.Hex2BytesFixed(keyGenInfo.GroupIdHex, 32)

		var groupId [32]byte
		copy(groupId[:], groupIdRaw)

		myIndex := big.NewInt(int64(partyInfo.Index))

		//groupId := m.keygenRequestGroups[requestId]
		////ind, err := m.getMyIndexInGroup(groupId)
		//m.storer.
		//	fmt.Printf("My index is %v\n", ind)
		//if err != nil {
		//	return err
		//}
		tx, err := m.instance.ReportGeneratedKey(m.signer, groupId, myIndex, genPubKey)
		// todo: to deal with: "error": "insufficient funds for gas * price + value: address 0x3600323b486F115CE127758ed84F26977628EeaA have (103000) want (3019200000000000)"}
		if err != nil {
			m.log.Error("Failed to report generated key",
				logger.Field{"groupId", groupId},
				logger.Field{"pubKey", result.Result},
				logger.Field{"error", err})
			return errors.Wrapf(err, "failed to report generated key %v for group %v", result.Result, groupId)
		}

		keyGenInfo.PubKeyReportedAt = time.Now()
		pubKeyHash := crypto.Keccak256Hash(genPubKey)
		keyGenInfo.PubKeyHashHex = pubKeyHash.Hex()

		err = m.storer.StoreKeygenRequestInfo(keyGenInfo)
		if err != nil {
			return errors.WithStack(err)
		}

		m.pendingReports[tx.Hash()] = &ReportKeyTx{
			groupId:            groupId,
			myIndex:            myIndex,
			generatedPublicKey: genPubKey,
		}
		if err != nil {
			return errors.WithStack(err)
		}
		m.log.Info("Reported generated public key",
			logger.Field{"publicKeyHex", result.Result},
			logger.Field{"txHashHex", tx.Hash().Hex()})
		delete(m.pendingKeygenRequests, requestId)
		return nil
	}
	m.log.Debug("Key hasn't been generated yet",
		logger.Field{"requestId", requestId},
		logger.Field{"requestType", result.RequestType},
		logger.Field{"statusStatus", result.RequestStatus})
	return nil
}

// todo: verify signature with third-party lib.
func (m *TaskManager) checkSignResult(signReqId string) error {
	signResult, err := m.mpcClient.Result(context.Background(), signReqId) // todo: add shared context to task manager
	m.log.Debug("Task-manager got sign result from mpc-server",
		logger.Field{"signResult", signResult})
	if err != nil {
		return err
	}
	if signResult.RequestStatus == "DONE" {
		var sig [65]byte
		sigBytes := common.FromHex(signResult.Result)
		//sigBytes := common.Hex2Bytes(signResult.Result[:])

		copy(sig[:], sigBytes)
		pendingTaskId, err := parsePendingRequestId(signReqId)
		if err != nil {
			return err
		}
		pendingSignReq := m.pendingSignRequests[signReqId]
		task := m.stakeTasks[pendingTaskId.taskId]
		var hashMismatchErr = errors.New("hash doesn't match")
		var wrongRequestNumberErr = errors.New("wrong request number")
		if pendingTaskId.requestNumber == 0 {
			m.log.Info("ExportHash have been signed from mpc-server=========step forward for ImportHash sign")
			// todo: verify signature with third-party lib.

			hashBytes, err := task.ExportTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hashBytes)
			if pendingSignReq.Hash != hashHex {
				fmt.Printf("%+v", errors.WithStack(hashMismatchErr))
				return hashMismatchErr
			}
			err = task.SetExportTxSig(sig)
			if err != nil {
				return err
			}

			// Delete signed export message
			delete(m.pendingSignRequests, signReqId)

			// Build next sign request
			pendingTaskId.requestNumber += 1
			nextPendingSignReq := new(core.SignRequest)
			nextPendingSignReq.RequestId = pendingTaskId.ToString()
			nextPendingSignReq.PublicKey = pendingSignReq.PublicKey
			nextPendingSignReq.ParticipantKeys = pendingSignReq.ParticipantKeys
			hashBytes, err = task.ImportTxHash()
			if err != nil {
				return errors.WithStack(err)
			}
			nextPendingSignReq.Hash = common.Bytes2Hex(hashBytes)

			err = m.mpcClient.Sign(context.Background(), nextPendingSignReq) // todo: add shared context to task manager
			m.log.Debug("Task-manager sent next sign request", logger.Field{"nextSignRequest", nextPendingSignReq})
			if err != nil {
				return errors.WithStack(err)
			}

			m.pendingSignRequests[nextPendingSignReq.RequestId] = nextPendingSignReq

		} else if pendingTaskId.requestNumber == 1 {
			// todo: verify signature with third-party lib.

			hashBytes, err := task.ImportTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hashBytes)
			if pendingSignReq.Hash != hashHex {
				fmt.Printf("%+v", errors.WithStack(hashMismatchErr))
				return hashMismatchErr
			}
			err = task.SetImportTxSig(sig)
			if err != nil {
				return err
			}
			delete(m.pendingSignRequests, signReqId)

			// Build next sign request
			pendingTaskId.requestNumber += 1
			nextPendingSignReq := new(core.SignRequest)
			nextPendingSignReq.RequestId = pendingTaskId.ToString()
			nextPendingSignReq.PublicKey = pendingSignReq.PublicKey
			nextPendingSignReq.ParticipantKeys = pendingSignReq.ParticipantKeys
			hashBytes, err = task.AddDelegatorTxHash()
			if err != nil {
				return errors.WithStack(err)
			}
			nextPendingSignReq.Hash = common.Bytes2Hex(hashBytes)

			err = m.mpcClient.Sign(context.Background(), nextPendingSignReq) // todo: add shared context to task manager
			m.log.Debug("Task-manager sent next sign request", logger.Field{"nextSignRequest", nextPendingSignReq})
			if err != nil {
				return errors.WithStack(err)
			}

			m.pendingSignRequests[nextPendingSignReq.RequestId] = nextPendingSignReq
		} else if pendingTaskId.requestNumber == 2 {
			hashBytes, err := task.AddDelegatorTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hashBytes)
			if pendingSignReq.Hash != hashHex {
				fmt.Printf("%+v", errors.WithStack(hashMismatchErr))
				return hashMismatchErr
			}
			err = task.SetAddDelegatorTxSig(sig)
			if err != nil {
				return err
			}
			delete(m.pendingSignRequests, signReqId)
			m.log.Info("Mpc-manager: Cool! All signings for a stake task all done.")
			ids, err := m.staker.IssueStakeTaskTxs(context.Background(), task)

			//err = doStake(task)
			if err != nil {
				m.log.Error("Failed to doStake",
					logger.Field{"error", err})
				return errors.WithStack(err)
			}
			m.log.Info("Mpc-manager: Cool! Success to add delegator!",
				logger.Field{"stakeTaske", task},
				logger.Field{"ids", ids})
		} else {
			fmt.Printf("%+v", errors.WithStack(hashMismatchErr))
			return wrongRequestNumberErr
		}
	}
	return nil
}

func (m *TaskManager) onKeygenRequestAdded(evt *contract.MpcCoordinatorKeygenRequestAdded) error {
	// Request MPC server
	//group, err := m.instance.GetGroup(nil, evt.GroupIdHex)
	//if err != nil {
	//	return err
	//}
	//var participantKeys []string
	//for _, k := range group.PartPubKeyHexs {
	//	pk := common.Bytes2Hex(k)
	//	participantKeys = append(participantKeys, pk)
	//}
	//
	//t := group.Threshold.Uint64()

	groupIdHex := common.Bytes2Hex(evt.GroupId[:])
	//groupInfo, ok := m.groupCache[groupIdHex]
	//if !ok {
	//	m.log.Error("Failed to get group from cache", logger.Field{"groupIdHex", groupIdHex})
	//	return pkgErrors.Errorf("Failed to get group from cache, groupIdHex: %q", groupIdHex)
	//}

	groupInfo, err := m.storer.LoadGroupInfo(groupIdHex)
	if err != nil {
		return errors.WithStack(err)
	}

	//request := &core.KeygenRequest{
	//	RequestId:       evt.Raw.TxHash.Hex(),
	//	ParticipantKeys: participantKeys,
	//
	//	Threshold: t,
	//}

	reqIdHex := evt.Raw.TxHash.Hex()
	partPubKeyHexs := groupInfo.PartPubKeyHexs
	request := &core.KeygenRequest{
		RequestId:       reqIdHex,
		ParticipantKeys: partPubKeyHexs,

		Threshold: groupInfo.Threshold,
	}

	err = m.mpcClient.Keygen(context.Background(), request) // todo: add shared context to task manager
	if err != nil {
		return errors.Wrapf(err, "failed to send keygen request to mpc-server")
	}
	m.pendingKeygenRequests[request.RequestId] = request
	//m.keygenRequestGroups[request.RequestId] = evt.GroupId

	keygenReqInfo := storage.KeygenRequestInfo{
		RequestIdHex:   reqIdHex,
		GroupIdHex:     groupIdHex,
		RequestAddedAt: time.Now(),
	}

	err = m.storer.StoreKeygenRequestInfo(&keygenReqInfo)
	if err != nil {
		return errors.WithStack(err)
	}
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

	//var groupIds [][32]byte
	//for groupIdHex, _ := range m.myIndicesInGroups {
	//	var groupId [32]byte
	//	groupIdRaw := common.Hex2BytesFixed(groupIdHex, 32)
	//	copy(groupId[:], groupIdRaw)
	//	groupIds = append(groupIds, groupId)
	//}

	groupIds, err := m.getMyGroupIds()
	if err != nil {
		return errors.WithStack(err)
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
	//var groupIds [][32]byte
	//for groupIdHex, _ := range m.myIndicesInGroups {
	//	var groupId [32]byte
	//	groupIdRaw := common.Hex2BytesFixed(groupIdHex, 32)
	//	copy(groupId[:], groupIdRaw)
	//	groupIds = append(groupIds, groupId)
	//}

	groupIds, err := m.getMyGroupIds()
	if err != nil {
		return errors.WithStack(err)
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
	//var pubkeys [][]byte
	//for _, pubKeyHex := range m.publicKeyCache {
	//	pk := common.Hex2Bytes(pubKeyHex)
	//	pubkeys = append(pubkeys, pk)
	//}

	pubkeys, err := m.getMyPubKeys()
	if err != nil {
		return errors.WithStack(err)
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
	//var pubkeys [][]byte
	//for _, pubKeyHex := range m.publicKeyCache {
	//	pk := common.Hex2Bytes(pubKeyHex)
	//	pubkeys = append(pubkeys, pk)
	//}

	pubkeys, err := m.getMyPubKeys()
	if err != nil {
		return errors.WithStack(err)
	}

	sub, err := m.listener.WatchStakeRequestStarted(nil, m.eventsStS, pubkeys)
	if err != nil {
		return err
	}
	m.subStS = sub
	return nil
}

func (m *TaskManager) onParticipantAdded(evt *contract.MpcCoordinatorParticipantAdded) error {
	// Store participant
	groupId := common.Bytes2Hex(evt.GroupId[:])
	p := storage.ParticipantInfo{
		PubKeyHashHex: m.myPubKeyHash.Hex(),
		PubKeyHex:     m.myPubKey,
		GroupIdHex:    groupId,
		Index:         evt.Index.Uint64(),
	}
	err := m.storer.StoreParticipantInfo(&p)
	if err != nil {
		return errors.Wrapf(err, "failed to store participant")
	}
	m.log.Debug("Stored a participant", logger.Field{"participant", p})

	// Store group
	group, err := m.instance.GetGroup(nil, evt.GroupId)
	if err != nil {
		m.log.Error("Failed to query group", logger.Field{"error", err})
		return errors.Wrapf(err, "failed to query group")
	}
	var participantKeys []string
	for _, k := range group.Participants {
		pk := common.Bytes2Hex(k)
		participantKeys = append(participantKeys, pk)
	}

	t := group.Threshold.Uint64()

	g := storage.GroupInfo{
		GroupIdHex:     groupId,
		PartPubKeyHexs: participantKeys,
		Threshold:      t,
	}
	err = m.storer.StoreGroupInfo(&g)
	if err != nil {
		return errors.Wrapf(err, "failed to store group")
	}
	m.log.Debug("Stored a group", logger.Field{"group", g})
	//m.groupCache[groupId] = &g

	// Subscribe event KeygenRequestAdded

	//m.myIndicesInGroups[groupId] = evt.Index
	err = m.subscribeKeygenRequestAdded()
	if err != nil {
		return errors.Wrapf(err, "failed to subscribe event KeygenRequestAdded")
	}

	// Subscribe event KeyGenerated
	err = m.subscribeKeyGenerated()
	if err != nil {
		return errors.Wrapf(err, "failed to subscribe event KeyGenerated")
	}

	return nil
}

func (m *TaskManager) onKeyGenerated(req *contract.MpcCoordinatorKeyGenerated) error {
	// Store generated public key
	pkHash := crypto.Keccak256Hash(req.PublicKey)
	pkHex := common.Bytes2Hex(req.PublicKey)
	groupId := common.Bytes2Hex(req.GroupId[:])
	pk := storage.GeneratedPubKeyInfo{
		PubKeyHashHex: pkHash.Hex(),
		PubKeyHex:     pkHex,
		GroupIdHex:    groupId,
	}
	err := m.storer.StoreGeneratedPubKeyInfo(&pk)
	if err != nil {
		return errors.Wrapf(err, "failed to store generated public key")
	}
	m.log.Debug("Stored a generated public ket", logger.Field{"genPubKey", pk})

	//m.publicKeyCache[hash] = pkHex

	// todo: only do the following if it's me added.

	// Subscribe event StakeRequestAdded
	err = m.subscribeStakeRequestAdded()
	if err != nil {
		return errors.Wrapf(err, "failed to subscribe event StakeRequestAdded")
	}

	// Subscribe event StakeRequestStarted
	err = m.subscribeStakeRequestStarted()
	if err != nil {
		return errors.Wrapf(err, "failed to subscribe event StakeRequestStarted")
	}

	return nil
}

// todo: store this event info
func (m *TaskManager) onStakeRequestAdded(req *contract.MpcCoordinatorStakeRequestAdded) error {
	//pubKeyHex := m.getPublicKey(req.PublicKey)
	//ind, err := m.getMyIndex(pubKeyHex)
	//if err != nil {
	//	return err
	//}

	ind, err := m.getMyIndex(req.PublicKey)
	if err != nil {
		return errors.WithStack(err)
	}

	tx, err := m.instance.JoinRequest(m.signer, req.RequestId, ind)
	if err != nil {
		fmt.Printf("Failed to joined stake request tx hash: %v\n", tx)
		return errors.WithStack(err)
	}
	j := &JoinTx{
		requestId: req.RequestId,
		myIndex:   ind,
	}
	m.pendingJoins[tx.Hash()] = j
	fmt.Printf("Joined stake request tx hash: %v\n", tx)
	return nil
}

func (m *TaskManager) removePendingJoin(requestId *big.Int) error {
	var txHash *common.Hash = nil
	for hash, req := range m.pendingJoins {
		if req.requestId.Cmp(requestId) == 0 {
			txHash = &hash
			break
		}
	}
	if txHash != nil {
		delete(m.pendingJoins, *txHash)
	}
	return nil
}

func (m *TaskManager) onStakeRequestStarted(req *contract.MpcCoordinatorStakeRequestStarted) error {
	m.removePendingJoin(req.RequestId)

	// Request MPC server
	//m.storer.LoadGeneratedPubKeyInfo()
	//
	//pubKey := m.getPublicKey(req.PublicKey)
	//myInd, err := m.getMyIndex(pubKey)
	//if err != nil {
	//	return err
	//}

	myInd, err := m.getMyIndex(req.PublicKey)
	if err != nil {
		return errors.WithStack(err)
	}

	var participating bool
	for _, ind := range req.ParticipantIndices {
		participating = participating || ind.Cmp(myInd) == 0
	}

	if !participating {
		// Not Participating, Ignore
		m.log.Info("Not participated to stake request", logger.Field{"stakeReqId", req.RequestId})
		return nil
	}

	nodeID, err := ids.ShortFromPrefixedString(req.NodeID, constants.NodeIDPrefix)

	if err != nil {
		return errors.WithStack(err)
	}

	pkHashHex := req.PublicKey.Hex()
	genPkInfo, err := m.storer.LoadGeneratedPubKeyInfo(pkHashHex)
	if err != nil {
		return errors.WithStack(err)
	}

	if genPkInfo == nil {
		return errors.New("No generated public key info found")
	}

	pkBytes := common.Hex2Bytes(genPkInfo.PubKeyHex)
	pk, err := unmarshalPubkey(pkBytes)
	if err != nil {
		return errors.WithStack(err)
	}
	cChainAddress := crypto.PubkeyToAddress(*pk)
	nonce, err := m.ethClient.NonceAt(context.Background(), cChainAddress, nil)

	if err != nil {
		return errors.WithStack(err)
	}

	baseFeeGwei := uint64(300) // TODO: It should be given by the contract
	if !req.Amount.IsUint64() || !req.StartTime.IsUint64() || !req.EndTime.IsUint64() {
		return errors.New("invalid uint64")
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
	pariticipantKeys, err := m.getPariticipantKeys(req.PublicKey, req.ParticipantIndices)
	if err != nil {
		return errors.WithStack(err)
	}

	m.log.Debug("Task-manager got participant public keys",
		logger.Field{"participantPubKeys", pariticipantKeys})

	normalized, err := myCrypto.NormalizePubKeys(pariticipantKeys)
	m.log.Debug("Task-manager normalized participant public keys",
		logger.Field{"normalizedParticipantPubKeys", normalized})
	if err != nil {
		return errors.Wrapf(err, "failed to normalized participant public keys: %v", pariticipantKeys)
	}

	genPkInfo, err = m.storer.LoadGeneratedPubKeyInfo(pkHashHex)
	if err != nil {
		return errors.WithStack(err)
	}

	reqId := PendingRequestId{taskId: taskId, requestNumber: 0}
	hash := common.Bytes2Hex(hashBytes)
	request := &core.SignRequest{
		RequestId: reqId.ToString(),
		PublicKey: genPkInfo.PubKeyHex,
		//ParticipantKeys: pariticipantKeys,
		ParticipantKeys: normalized,
		Hash:            hash,
	}
	err = m.mpcClient.Sign(context.Background(), request) // todo: add shared context to task manager
	if err != nil {
		return errors.WithStack(err)
	}
	m.pendingSignRequests[request.RequestId] = request

	//if pkHashHex, ok := m.publicKeyCache[req.PublicKey]; ok {
	//	pkBytes := common.Hex2Bytes(pkHashHex)
	//
	//	pk, err := unmarshalPubkey(pkBytes)
	//	if err != nil {
	//		return err
	//	}
	//	cChainAddress := crypto.PubkeyToAddress(*pk)
	//	nonce, err := m.ethClient.NonceAt(context.Background(), cChainAddress, nil)
	//
	//	if err != nil {
	//		return err
	//	}
	//
	//	var invalidUint64Err = errors.New("invalid uint64")
	//	baseFeeGwei := uint64(300) // TODO: It should be given by the contract
	//	if !req.Amount.IsUint64() || !req.StartTime.IsUint64() || !req.EndTime.IsUint64() {
	//		return invalidUint64Err
	//	}
	//	task, err := NewStakeTask(m.networkContext, *pk, nonce, nodeID, req.Amount.Uint64(), req.StartTime.Uint64(), req.EndTime.Uint64(), baseFeeGwei)
	//	if err != nil {
	//		return err
	//	}
	//	taskId := req.Raw.TxHash.Hex()
	//	m.stakeTasks[taskId] = task
	//	hashBytes, err := task.ExportTxHash()
	//	if err != nil {
	//		return err
	//	}
	//	pariticipantKeys, err := m.getPariticipantKeys(pubKey, req.ParticipantIndices)
	//	if err != nil {
	//		return errors.WithStack(err)
	//	}
	//
	//	m.log.Debug("Task-manager got participant public keys",
	//		logger.Field{"participantPubKeys", pariticipantKeys})
	//
	//	normalized, err := myCrypto.NormalizePubKeys(pariticipantKeys)
	//	m.log.Debug("Task-manager normalized participant public keys",
	//		logger.Field{"normalizedParticipantPubKeys", normalized})
	//	if err != nil {
	//		return errors.Wrapf(err, "failed to normalized participant public keys: %v", pariticipantKeys)
	//	}
	//
	//	reqId := PendingRequestId{taskId: taskId, requestNumber: 0}
	//	hash := common.Bytes2Hex(hashBytes)
	//	request := &core.SignRequest{
	//		RequestId: reqId.ToString(),
	//		PublicKey: pubKey,
	//		//ParticipantKeys: pariticipantKeys,
	//		ParticipantKeys: normalized,
	//		Hash:            hash,
	//	}
	//	err = m.mpcClient.Sign(context.Background(), request) // todo: add shared context to task manager
	//	if err != nil {
	//		return errors.WithStack(err)
	//	}
	//	m.pendingSignRequests[request.RequestId] = request
	//}
	return nil
}

func (m *TaskManager) requestKeygen(req *contract.MpcCoordinatorKeygenRequestAdded) error {
	/*
		m.mpcClient.Keygen(core.KeygenRequest{RequestId: req.Raw.TxHash.Hex(), })
		ParticipantKeys
		res, err := m.instance.GetGroup(nil, req.GroupIdHex)
		if err != nil {
			return err
		}
		t := res.Threshold.String()
		id := req.Raw.TxHash.Hex()
		pubKeys := ""
		for i, pk := range res.PartPubKeyHexs {
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

//func (m *TaskManager) getPublicKey(topic common.Hash) string {
//	return m.publicKeyCache[topic]
//}

//func (m *TaskManager) getMyIndex(publicKey string) (*big.Int, error) {
//
//	k := common.Hex2Bytes(publicKey)
//
//	//inf, err := m.instance.GetKey(nil, k)
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	group, err := m.instance.GetGroup(nil, inf.GroupId)
//	if err != nil {
//		return nil, err
//	}
//
//	groupId := common.Bytes2Hex(inf.GroupId[:])
//	g, ok := m.groupCache[groupId]
//	if !ok {
//		m.log.Error("Failed to get group from cache", logger.Field{"groupId", groupId})
//		return nil, pkgErrors.Errorf("Failed to get group from cache, groupId: %q", groupId)
//	}
//	for i, pkBytes := range group.Participants {
//		pk := common.Bytes2Hex(pkBytes)
//		if m.myPubKey == pk {
//			return big.NewInt(int64(i) + 1), nil
//		}
//	}
//	return nil, errors.New("not a member of the group")
//}

//func (m *TaskManager) getMyIndexInGroup(groupId [32]byte) (*big.Int, error) {
//
//	group, err := m.instance.GetGroup(nil, groupId)
//	if err != nil {
//		return nil, err
//	}
//	for i, pkBytes := range group.Participants {
//		pk := common.Bytes2Hex(pkBytes)
//		if m.myPubKey == pk {
//			return big.NewInt(int64(i) + 1), nil
//		}
//	}
//	return nil, errors.New("not a member of the group")
//}

//func (m *TaskManager) getPariticipantKeys(publicKey string, indices []*big.Int) ([]string, error) {
//
//	k := common.Hex2Bytes(publicKey)
//
//	inf, err := m.instance.GetKey(nil, k)
//	if err != nil {
//		return nil, err
//	}
//
//	group, err := m.instance.GetGroup(nil, inf.GroupId)
//	if err != nil {
//		return nil, err
//	}
//	var out []string
//	for _, ind := range indices {
//		k := group.Participants[ind.Uint64()-1]
//		pk := common.Bytes2Hex(k)
//		out = append(out, pk)
//	}
//	return out, nil
//}

func (m *TaskManager) getMyIndex(genPubKeyHash common.Hash) (*big.Int, error) {
	genPubKeyInfo, err := m.storer.LoadGeneratedPubKeyInfo(genPubKeyHash.Hex())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	partInfo, err := m.storer.LoadParticipantInfo(m.myPubKeyHash.Hex(), genPubKeyInfo.GroupIdHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return big.NewInt(int64(partInfo.Index)), nil
}

func (m *TaskManager) getMyGroupIds() ([][32]byte, error) {
	partInfos, err := m.storer.LoadParticipantInfos(m.myPubKeyHash.Hex())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var groupIds [][32]byte

	for _, partInfo := range partInfos {
		var groupId [32]byte
		groupIdRaw := common.Hex2BytesFixed(partInfo.GroupIdHex, 32)
		copy(groupId[:], groupIdRaw)
		groupIds = append(groupIds, groupId)
	}

	return groupIds, nil
}

func (m *TaskManager) getMyPubKeys() ([][]byte, error) {
	partyInfos, err := m.storer.LoadParticipantInfos(m.myPubKeyHash.Hex())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var groupIdHexs []string
	for _, partyInfo := range partyInfos {
		groupIdHexs = append(groupIdHexs, partyInfo.GroupIdHex)
	}
	if len(groupIdHexs) == 0 {
		return nil, errors.New("found no group")
	}

	genPubKeyInfos, err := m.storer.LoadGeneratedPubKeyInfos(groupIdHexs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(genPubKeyInfos) == 0 {
		return nil, errors.New("found no generated public key")
	}

	var genPubKeyBytes [][]byte
	for _, genPubKeyInfo := range genPubKeyInfos {
		pubKeyBytes := common.Hex2Bytes(genPubKeyInfo.PubKeyHex)
		genPubKeyBytes = append(genPubKeyBytes, pubKeyBytes)
	}

	return genPubKeyBytes, nil
}

func (m *TaskManager) getPariticipantKeys(genPubKeyHash common.Hash, indices []*big.Int) ([]string, error) {
	genPkInfo, err := m.storer.LoadGeneratedPubKeyInfo(genPubKeyHash.Hex())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	groupInfo, err := m.storer.LoadGroupInfo(genPkInfo.GroupIdHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var partPubKeyHexs []string
	for _, ind := range indices {
		partPubKeyHex := groupInfo.PartPubKeyHexs[ind.Uint64()-1]
		partPubKeyHexs = append(partPubKeyHexs, partPubKeyHex)
	}
	return partPubKeyHexs, nil
}

func unmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	if pub[0] == 4 {
		x, y := elliptic.Unmarshal(crypto.S256(), pub)
		if x == nil {
			return nil, errors.Errorf("invalid secp256k1 public key %v", common.Bytes2Hex(pub))
		}
		return &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}, nil
	} else {
		x, y := secp256k1.DecompressPubkey(pub)
		if x == nil {
			return nil, errors.Errorf("invalid secp256k1 public key %v", common.Bytes2Hex(pub))
		}
		return &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}, nil
	}
}

func marshalPubkey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
}
