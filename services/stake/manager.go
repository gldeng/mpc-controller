package stake

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	ctlPk "github.com/avalido/mpc-controller"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"golang.org/x/sync/errgroup"

	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	avaEthclient "github.com/ava-labs/coreth/ethclient"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

type Manager struct {
	ctx    context.Context
	config config.Config
	log    logger.Logger
	staker *Staker

	mpcControllerId string
	networkContext  core.NetworkContext

	stakeTasks map[string]*StakeTask

	pendingSignRequests map[string]*core.SignRequest

	pendingJoins map[common.Hash]*JoinTx

	avaEthclient    avaEthclient.Client
	myAddr          ids.ShortID
	coordinatorAddr common.Address
	eventsStS       chan *contract.MpcManagerStakeRequestStarted
	eventsStA       chan *contract.MpcManagerStakeRequestAdded

	ctlPk.StorerGetGroupIds
	ctlPk.TransactorJoinRequest
	ctlPk.StorerGetPubKeys
	ctlPk.WatcherStakeRequestAdded
	ctlPk.WatcherStakeRequestStarted
	ctlPk.CallerGetGroup
	ctlPk.StorerLoadGeneratedPubKeyInfo

	ctlPk.EthClientTransactionReceipt
	ctlPk.EthClientNonceAt
	ctlPk.EthClientBalanceAt

	secpFactory   avaCrypto.FactorySECP256K1R
	chSigReceived chan *SignatureReceived
	mpcClient     core.MpcClient
	signer        *bind.TransactOpts
	myPubKey      string
	myPubKeyHash  common.Hash

	subStA event.Subscription
	subStS event.Subscription

	ctlPk.StorerGetParticipantIndex
	ctlPk.StorerGetPariticipantKeys
}

func NewTaskManager(ctx context.Context, log logger.Logger, config config.Config, staker *Staker) (*Manager, error) {
	privKey := config.ControllerKey()
	pubKeyBytes := marshalPubkey(&privKey.PublicKey)[1:]
	pubKeyHex := common.Bytes2Hex(pubKeyBytes)
	pubKeyHash := crypto.Keccak256Hash(pubKeyBytes)
	log.Debug("parsed task manager key info",
		logger.Field{"mpcControllerId", config.ControllerId()},
		logger.Field{"pubKey", pubKeyHex},
		logger.Field{"pubKeyTopic", pubKeyHash})
	m := &Manager{
		ctx:                 ctx,
		config:              config,
		log:                 log,
		staker:              staker,
		networkContext:      *config.NetworkContext(),
		mpcClient:           config.MpcClient(),
		signer:              config.ControllerSigner(),
		myPubKey:            pubKeyHex,
		myPubKeyHash:        pubKeyHash,
		coordinatorAddr:     *config.CoordinatorAddress(),
		stakeTasks:          make(map[string]*StakeTask),
		pendingSignRequests: make(map[string]*core.SignRequest),
		pendingJoins:        make(map[common.Hash]*JoinTx),
	}

	m.chSigReceived = make(chan *SignatureReceived)
	m.eventsStA = make(chan *contract.MpcManagerStakeRequestAdded)
	m.eventsStS = make(chan *contract.MpcManagerStakeRequestStarted)
	m.secpFactory = avaCrypto.FactorySECP256K1R{}
	return m, nil
}

// todo: logic to quit for loop

func (m *Manager) Start(ctx context.Context) error {
	// Initiate event subscription with mpc-coordinator
	g, ctx := errgroup.WithContext(m.ctx)

	// Do the heavy core service
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil

			case evt, ok := <-m.eventsStA:
				if !ok {
					m.log.Debug("Retrieve nothing from StakeRequestAdded event channel")
					break
				}

				m.log.Info("Received StakeRequestAdded event", logger.Field{"event", evt})

				err := m.onStakeRequestAdded(evt)
				if err != nil {
					m.log.Error("Failed to respond to StakeRequestAdded event", []logger.Field{
						{"event", evt},
						{"error", err}}...)
				}

			case evt, ok := <-m.eventsStS:
				if !ok {
					m.log.Debug("Retrieve nothing from StakeRequestStarted event channel")
					break
				}

				m.log.Info("Received StakeRequestStarted event", logger.Field{"event", evt})

				//// Wait until the corresponding key has been generated
				//<-time.After(time.Second * 20)

				err := m.onStakeRequestStarted(evt)
				if err != nil {
					m.log.Error("Failed to respond to StakeRequestStarted event", logger.Field{"error", err})
				}

			case <-time.After(1 * time.Second):
				err := m.tick()
				if err != nil {
					m.log.Error("Got an tick error", logger.Field{"error", err})
				}
			}
		}
	})

	fmt.Printf("%v service started.\n", m.config.ControllerId())
	if err := g.Wait(); err != nil {
		return errors.WithStack(err)
	}

	// Release supportive resources after no further consumption
	//m.ethWsClient.Close()

	fmt.Printf("%v service closed.\n", m.config.ControllerId())
	return nil
}

func (m *Manager) tick() error {
	err := m.checkPendingJoins()
	if err != nil {
		return err
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

func (m *Manager) checkPendingJoins() error {
	var done []common.Hash
	var retry []common.Hash
	for txHash, _ := range m.pendingJoins {
		rcp, err := m.TransactionReceipt(m.ctx, txHash)
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
		tx, err := m.JoinRequest(m.signer, requestId, myIndex)
		m.pendingJoins[tx.Hash()] = &JoinTx{
			requestId: requestId,
			myIndex:   myIndex,
		}
		if err != nil {
			m.log.Error("Failed to join request", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		m.log.Info("Retry join request.", []logger.Field{
			{"reqId", requestId},
			{"myIndex", myIndex},
			{"txHash", tx.Hash()}}...)
	}
	for _, txHash := range sampledRetry {
		delete(m.pendingJoins, txHash)
	}
	for _, txHash := range done {
		delete(m.pendingJoins, txHash)
	}
	return nil
}

// todo: verify signature with third-party lib.
func (m *Manager) checkSignResult(signReqId string) error {
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
			// todo: verify signature with third-party lib.

			hashBytes, err := task.ExportTxHash()
			if err != nil {
				return err
			}
			hashHex := common.Bytes2Hex(hashBytes)
			if pendingSignReq.Hash != hashHex {
				m.log.Error("Hash doesn't match")
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
				m.log.Error("Hash doesn't match")
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
				m.log.Error("Hash doesn't match")
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
			m.log.Error("Hash doesn't match")
			return wrongRequestNumberErr
		}
	}
	return nil
}

func (m *Manager) subscribeStakeRequestAdded() error {
	if m.subStA != nil {
		m.subStA.Unsubscribe()
		m.subStA = nil
	}
	pubkeys, err := m.GetPubKeys(m.myPubKeyHash.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	sub, err := m.WatchStakeRequestAdded(pubkeys)
	if err != nil {
		return err
	}
	m.subStA = sub
	return nil
}

func (m *Manager) subscribeStakeRequestStarted() error {
	if m.subStS != nil {
		m.subStS.Unsubscribe()
		m.subStS = nil
	}
	pubkeys, err := m.GetPubKeys(m.myPubKeyHash.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	sub, err := m.WatchStakeRequestStarted(pubkeys)
	if err != nil {
		return err
	}
	m.subStS = sub
	return nil
}

// todo: store this event info
func (m *Manager) onStakeRequestAdded(req *contract.MpcManagerStakeRequestAdded) error {
	ind, err := m.GetIndex(m.myPubKeyHash.Hex(), req.PublicKey.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	tx, err := m.JoinRequest(m.signer, req.RequestId, ind)
	if err != nil {
		m.log.Error("Failed to join stake request", []logger.Field{{"error", err}, {"tx", tx}}...)
		return errors.WithStack(err)
	}
	j := &JoinTx{
		requestId: req.RequestId,
		myIndex:   ind,
	}
	m.pendingJoins[tx.Hash()] = j
	return nil
}

func (m *Manager) removePendingJoin(requestId *big.Int) error {
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

func (m *Manager) onStakeRequestStarted(req *contract.MpcManagerStakeRequestStarted) error {
	m.removePendingJoin(req.RequestId)

	myInd, err := m.GetIndex(m.myPubKeyHash.Hex(), req.PublicKey.Hex())
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
	genPkInfo, err := m.LoadGeneratedPubKeyInfo(pkHashHex)
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

	nonce, err := m.NonceAt(m.ctx, *address, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	bl, _ := m.BalanceAt(m.ctx, *address, nil)
	m.log.Debug("$$$$$$$$$C Balance of C-Chain address before export", []logger.Field{
		{"address", *address},
		{"balance", bl.Uint64()}}...)

	baseFeeGwei := uint64(300) // TODO: It should be given by the contract

	nAVAXAmount := new(big.Int).Div(req.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !req.StartTime.IsUint64() || !req.EndTime.IsUint64() {
		return errors.New("invalid uint64")
	}
	task, err := NewStakeTask(m.networkContext, *pubkey, nonce, nodeID, nAVAXAmount.Uint64(), req.StartTime.Uint64(), req.EndTime.Uint64(), baseFeeGwei)
	if err != nil {
		return errors.WithStack(err)
	}
	taskId := req.Raw.TxHash.Hex()
	m.stakeTasks[taskId] = task
	hashBytes, err := task.ExportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}
	pariticipantKeys, err := m.GetPariticipantKeys(req.PublicKey.Hex(), req.ParticipantIndices)
	if err != nil {
		return errors.WithStack(err)
	}

	m.log.Debug("Task-manager got participant public keys", logger.Field{"participantPubKeys", pariticipantKeys})

	normalized, err := myCrypto.NormalizePubKeys(pariticipantKeys)
	m.log.Debug("Task-manager normalized participant public keys",
		logger.Field{"normalizedParticipantPubKeys", normalized})
	if err != nil {
		return errors.Wrapf(err, "failed to normalized participant public keys: %v", pariticipantKeys)
	}

	genPkInfo, err = m.LoadGeneratedPubKeyInfo(pkHashHex)
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

func marshalPubkey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
}
