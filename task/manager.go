package task

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/sync/errgroup"

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

// TODO: Add startTime to handle timeouts

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
	ctx    context.Context
	config config.Config
	log    logger.Logger
	storer storage.Storer
	staker *Staker

	mpcControllerId string
	networkContext  core.NetworkContext

	stakeTasks map[string]*StakeTask

	pendingSignRequests   map[string]*core.SignRequest
	pendingKeygenRequests map[string]*core.KeygenRequest

	pendingReports map[common.Hash]*ReportKeyTx

	pendingJoins map[common.Hash]*JoinTx

	avaEthclient        avaEthclient.Client
	myAddr              ids.ShortID
	coordinatorAddr     common.Address
	ethWsClient         *ethclient.Client
	cChainClient        evm.Client
	eventsKA            chan *contract.MpcCoordinatorKeygenRequestAdded
	eventsStS           chan *contract.MpcCoordinatorStakeRequestStarted
	eventsStA           chan *contract.MpcCoordinatorStakeRequestAdded
	rebuildListener     chan struct{}
	rebuildListenerDone chan struct{}
	listener            *contract.MpcCoordinator
	instance            *contract.MpcCoordinator
	ethRpcClient        *ethclient.Client
	secpFactory         avaCrypto.FactorySECP256K1R
	chSigReceived       chan *SignatureReceived
	mpcClient           core.MpcClient
	signer              *bind.TransactOpts
	myPubKey            string
	myPubKeyHash        common.Hash
	eventsPA            chan *contract.MpcCoordinatorParticipantAdded
	subPA               event.Subscription
	subKA               event.Subscription
	subStA              event.Subscription
	subStS              event.Subscription
	subKG               event.Subscription
	eventsKG            chan *contract.MpcCoordinatorKeyGenerated
}

func NewTaskManager(ctx context.Context, log logger.Logger, config config.Config, storer storage.Storer, staker *Staker) (*TaskManager, error) {
	privKey := config.ControllerKey()
	pubKeyBytes := marshalPubkey(&privKey.PublicKey)[1:]
	pubKeyHex := common.Bytes2Hex(pubKeyBytes)
	pubKeyHash := crypto.Keccak256Hash(pubKeyBytes)
	log.Debug("parsed task manager key info",
		logger.Field{"mpcControllerId", config.ControllerId()},
		logger.Field{"pubKey", pubKeyHex},
		logger.Field{"pubKeyTopic", pubKeyHash})
	m := &TaskManager{
		config:                config,
		log:                   log,
		staker:                staker,
		storer:                storer,
		networkContext:        *config.NetworkContext(),
		mpcClient:             config.MpcClient(),
		signer:                config.ControllerSigner(),
		myPubKey:              pubKeyHex,
		myPubKeyHash:          pubKeyHash,
		coordinatorAddr:       *config.CoordinatorAddress(),
		stakeTasks:            make(map[string]*StakeTask),
		pendingSignRequests:   make(map[string]*core.SignRequest),
		pendingKeygenRequests: make(map[string]*core.KeygenRequest),
		pendingReports:        make(map[common.Hash]*ReportKeyTx),
		pendingJoins:          make(map[common.Hash]*JoinTx),
	}

	m.listener = config.CoordinatorBoundListener()
	m.instance = config.CoordinatorBoundInstance()
	m.ethWsClient = config.EthWsClient()
	m.ethRpcClient = config.EthRpcClient()
	m.chSigReceived = make(chan *SignatureReceived)
	m.rebuildListener = make(chan struct{})
	m.rebuildListenerDone = make(chan struct{})
	m.eventsPA = make(chan *contract.MpcCoordinatorParticipantAdded)
	m.eventsKA = make(chan *contract.MpcCoordinatorKeygenRequestAdded)
	m.eventsKG = make(chan *contract.MpcCoordinatorKeyGenerated)
	m.eventsStA = make(chan *contract.MpcCoordinatorStakeRequestAdded)
	m.eventsStS = make(chan *contract.MpcCoordinatorStakeRequestStarted)
	m.secpFactory = avaCrypto.FactorySECP256K1R{}
	return m, nil
}

// todo: logic to quit for loop

func (m *TaskManager) Start() error {
	// Initiate event subscription with mpc-coordinator
	err := m.subscribeParticipantAdded()
	if err != nil {
		return errors.WithStack(err)
	}
	err = m.subscribeKeygenRequestAdded()
	if err != nil {
		return errors.WithStack(err)
	}
	err = m.subscribeKeyGenerated()
	if err != nil {
		return errors.WithStack(err)
	}

	g, ctx := errgroup.WithContext(m.ctx)
	// Continuously check websocket connection with Avalanche network (C-Chain).
	// Try to rebuild the websocket and coordinator if the network connection is down.
	// Also resubscribe three events, namely: ParticipantAdded, KeygenRequestAdded and KeyGenerated.
	g.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return nil
			case <-ticker.C:
				_, err := m.ethWsClient.NetworkID(ctx)
				if err != nil {
					m.log.Error("Failed to connect websocket", logger.Field{"error", err})
					m.rebuildListener <- struct{}{}
				rebuild:
					for {
						select {
						case <-ctx.Done():
							ticker.Stop()
							return nil
						case <-m.rebuildListenerDone:
							break rebuild
						}
					}
				}
			}
		}
	})
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-m.rebuildListener:
				wsClient, listener, err := m.config.CoordinatorBoundListenerRebuild(m.log, ctx)
				if err != nil {
					m.log.Error("Failed to rebuild coordinator listener", logger.Field{"error", err})
					break
				}
				// todo: pay attention to data race condition
				m.ethWsClient = wsClient
				m.instance = listener

				err = m.subscribeParticipantAdded()
				if err != nil {
					m.log.Error("Failed to resubscribe event ParticipantAdded", logger.Field{"error", err})
				}
				err = m.subscribeKeygenRequestAdded()
				if err != nil {
					m.log.Error("Failed to resubscribe event KeygenRequestAdded", logger.Field{"error", err})
				}
				err = m.subscribeKeyGenerated()
				if err != nil {
					m.log.Error("Failed to resubscribe event KeyGenerated", logger.Field{"error", err})
				}

				m.rebuildListenerDone <- struct{}{}

				m.log.Debug("Websocket client and coordinator listener rebuilt, plus events resubscribed")
			}
		}
	})

	// Do the heavy core service
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
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
	})

	fmt.Printf("%v service started.\n", m.config.ControllerId())
	if err := g.Wait(); err != nil {
		return errors.WithStack(err)
	}

	// Release supportive resources after no further consumption
	m.ethWsClient.Close()

	fmt.Printf("%v service closed.\n", m.config.ControllerId())
	return nil
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
		rcp, err := m.ethRpcClient.TransactionReceipt(m.ctx, txHash)
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
		rcp, err := m.ethRpcClient.TransactionReceipt(m.ctx, txHash)
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
	// Query result from mpc-server
	result, err := m.mpcClient.Result(m.ctx, requestId) // todo: add shared context to task manager
	if err != nil {
		return errors.WithStack(err)
	}

	if result.RequestStatus != "DONE" {
		m.log.Debug("Key hasn't been generated yet",
			[]logger.Field{{"reqId", requestId}, {"reqStatus", result.RequestStatus}}...)
		return nil
	}

	// Deal with crypto values
	dnmGenPubKeyBytes, err := myCrypto.DenormalizePubKeyFromHex(result.Result) // for Ethereum compatibility
	if err != nil {
		return errors.WithStack(err)
	}
	pubKeyHash := crypto.Keccak256Hash(dnmGenPubKeyBytes) // digest to identify generated public key

	// Load pre-stored corresponding participant and keygen request information
	keyGenInfo, err := m.storer.LoadKeygenRequestInfo(requestId) // keygen request info
	if err != nil {
		return errors.WithStack(err)
	}
	partyInfo, err := m.storer.LoadParticipantInfo(m.myPubKeyHash.Hex(), keyGenInfo.GroupIdHex) // participant info
	if err != nil {
		return errors.WithStack(err)
	}

	groupIdRaw := common.Hex2BytesFixed(keyGenInfo.GroupIdHex, 32)
	var groupId [32]byte
	copy(groupId[:], groupIdRaw)
	myIndex := big.NewInt(int64(partyInfo.Index))

	// Locally store the generated public key
	pk := storage.GeneratedPubKeyInfo{
		PubKeyHashHex: pubKeyHash.Hex(),
		PubKeyHex:     result.Result,
		GroupIdHex:    keyGenInfo.GroupIdHex,
	}
	err = m.storer.StoreGeneratedPubKeyInfo(&pk)
	if err != nil {
		return errors.Wrapf(err, "failed to store generated public key")
	}

	// Report the generated public key, in denormalized format due to Ethereum compatibility
	// Todo: establish a strategy to deal with "insufficient fund" error, maybe check account balance before report
	tx, err := m.instance.ReportGeneratedKey(m.signer, groupId, myIndex, dnmGenPubKeyBytes)
	if err != nil {
		m.log.Error("Failed to report public key", logger.Field{"error", err})
		return errors.Wrap(err, "failed to report generated key")
	}

	// Locally update keygen request information
	keyGenInfo.PubKeyReportedAt = time.Now()
	keyGenInfo.PubKeyHashHex = pubKeyHash.Hex()
	err = m.storer.StoreKeygenRequestInfo(keyGenInfo)
	if err != nil {
		return errors.WithStack(err)
	}

	m.pendingReports[tx.Hash()] = &ReportKeyTx{
		groupId:            groupId,
		myIndex:            myIndex,
		generatedPublicKey: dnmGenPubKeyBytes,
	}
	delete(m.pendingKeygenRequests, requestId)

	addr, _ := myCrypto.PubKeyHexToAddress(result.Result) // for debug
	m.log.Info("Generated and reported public key",
		[]logger.Field{{"ethAddress", addr}, {"generatedPubkey", result.Result},
			{"reportedPubkey", common.Bytes2Hex(dnmGenPubKeyBytes)}}...)
	return nil
}

// todo: verify signature with third-party lib.
func (m *TaskManager) checkSignResult(signReqId string) error {
	signResult, err := m.mpcClient.Result(m.ctx, signReqId) // todo: add shared context to task manager
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

			err = m.mpcClient.Sign(m.ctx, nextPendingSignReq) // todo: add shared context to task manager
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

			err = m.mpcClient.Sign(m.ctx, nextPendingSignReq) // todo: add shared context to task manager
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

			ids, err := m.staker.IssueStakeTaskTxs(m.ctx, task)

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
	groupIdHex := common.Bytes2Hex(evt.GroupId[:])

	groupInfo, err := m.storer.LoadGroupInfo(groupIdHex)
	if err != nil {
		return errors.WithStack(err)
	}

	reqIdHex := evt.Raw.TxHash.Hex()
	partPubKeyHexs := groupInfo.PartPubKeyHexs
	request := &core.KeygenRequest{
		RequestId:       reqIdHex,
		ParticipantKeys: partPubKeyHexs,

		Threshold: groupInfo.Threshold,
	}

	err = m.mpcClient.Keygen(m.ctx, request) // todo: add shared context to task manager
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
	// todo: only do the following if it's me added.

	// Subscribe event StakeRequestAdded
	err := m.subscribeStakeRequestAdded()
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

	pubkey, err := myCrypto.UnmarshalPubKeyHex(genPkInfo.PubKeyHex)
	if err != nil {
		return errors.WithStack(err)
	}

	address := myCrypto.PubkeyToAddresse(pubkey)

	nonce, err := m.ethRpcClient.NonceAt(m.ctx, *address, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	bl, _ := m.ethRpcClient.BalanceAt(m.ctx, *address, nil)
	m.log.Debug("$$$$$$$$$C Balance of C-Chain address before export", []logger.Field{{"address", *address}, {"balance", bl.Uint64()}}...)

	baseFeeGwei := uint64(300) // TODO: It should be given by the contract

	nAVAXAmount := new(big.Int).Div(req.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !req.StartTime.IsUint64() || !req.EndTime.IsUint64() {
		return errors.New("invalid uint64")
	}
	task, err := NewStakeTask(m.networkContext, *pubkey, nonce, nodeID, nAVAXAmount.Uint64(), req.StartTime.Uint64(), req.EndTime.Uint64(), baseFeeGwei)
	if err != nil {
		return err
	}
	// todo: remove this
	spew.Dump(task)
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
	err = m.mpcClient.Sign(m.ctx, request) // todo: add shared context to task manager
	if err != nil {
		return errors.WithStack(err)
	}
	m.pendingSignRequests[request.RequestId] = request
	return nil
}

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
		dnmGenPubKeyBytes, err := myCrypto.DenormalizePubKeyFromHex(genPubKeyInfo.PubKeyHex) // for Ethereum compatibility
		if err != nil {
			return nil, errors.WithStack(err)
		}

		//pubKeyBytes := common.Hex2Bytes(genPubKeyInfo.PubKeyHex)
		genPubKeyBytes = append(genPubKeyBytes, dnmGenPubKeyBytes)
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

func marshalPubkey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
}
