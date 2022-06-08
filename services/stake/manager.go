package stake

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	ctlPk "github.com/avalido/mpc-controller"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/services"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"time"
)

const sep = "-"

type JoinTx struct {
	requestId *big.Int
	myIndex   *big.Int
}

type PendingRequestId struct {
	taskId        string
	requestNumber uint8
}

type Manager struct {
	logger.Logger

	*Staker

	core.MpcClient
	core.NetworkContext

	ctlPk.TransactorJoinRequest
	ctlPk.WatcherStakeRequestAdded
	ctlPk.WatcherStakeRequestStarted

	ctlPk.EthClientTransactionReceipt
	ctlPk.EthClientNonceAt
	ctlPk.EthClientBalanceAt

	ctlPk.StorerGetParticipantIndex
	ctlPk.StorerGetPariticipantKeys
	ctlPk.StorerGetPubKeys
	ctlPk.StorerLoadGeneratedPubKeyInfo

	Signer     *bind.TransactOpts
	PubKeyHash common.Hash

	stakeTasks map[string]*StakeTask

	pendingSignRequests map[string]*core.SignRequest
	pendingJoins        map[common.Hash]*JoinTx

	stakeRequestAddedEvt   chan *contract.MpcManagerStakeRequestAdded
	stakeRequestStartedEvt chan *contract.MpcManagerStakeRequestStarted
}

func (m *Manager) Start(ctx context.Context) error {
	// Assign unexported fileds
	m.stakeTasks = make(map[string]*StakeTask)
	m.pendingSignRequests = make(map[string]*core.SignRequest)
	m.pendingJoins = make(map[common.Hash]*JoinTx)
	m.stakeRequestAddedEvt = make(chan *contract.MpcManagerStakeRequestAdded)
	m.stakeRequestStartedEvt = make(chan *contract.MpcManagerStakeRequestStarted)

	// Watch StakeRequestAdded and StakeRequestStarted events
	go func() {
		err := m.watchStakeRequest(ctx)
		m.ErrorOnError(err, "Got an error to watch state request events")
	}()

	// Actions upon events happening
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-m.stakeRequestAddedEvt:
			err := m.onStakeRequestAdded(ctx, evt)
			m.ErrorOnError(err, "Failed to process StakeRequestAdded event")
		case evt := <-m.stakeRequestStartedEvt:
			err := m.onStakeRequestStarted(ctx, evt)
			m.ErrorOnError(err, "Failed to process StakeRequestStarted event")
		case <-time.After(1 * time.Second):
			err := m.tick(ctx)
			if err != nil {
				m.Error("Got an tick error", logger.Field{"error", err})
			}
		}
	}
}

func (m *Manager) tick(ctx context.Context) error {
	err := m.checkPendingJoins(ctx)
	if err != nil {
		return err
	}
	for requestId, _ := range m.pendingSignRequests {
		err := m.checkSignResult(ctx, requestId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) checkPendingJoins(ctx context.Context) error {
	var done []common.Hash
	var retry []common.Hash
	for txHash, _ := range m.pendingJoins {
		rcp, err := m.TransactionReceipt(ctx, txHash)
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
	sampledRetry := services.Sample(retry)

	for _, txHash := range sampledRetry {
		req := m.pendingJoins[txHash]
		requestId, myIndex := req.requestId, req.myIndex
		tx, err := m.JoinRequest(ctx, m.Signer, requestId, myIndex)
		m.pendingJoins[tx.Hash()] = &JoinTx{
			requestId: requestId,
			myIndex:   myIndex,
		}
		if err != nil {
			m.Error("Failed to join request", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		m.Info("Retry join request.", []logger.Field{
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
func (m *Manager) checkSignResult(ctx context.Context, signReqId string) error {
	signResult, err := m.MpcClient.Result(ctx, signReqId) // todo: add shared context to task manager
	m.Debug("Task-manager got sign result from mpc-server",
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
				m.Error("Hash doesn't match")
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

			err = m.MpcClient.Sign(ctx, nextPendingSignReq) // todo: add shared context to task manager
			m.Debug("Task-manager sent next sign request", logger.Field{"nextSignRequest", nextPendingSignReq})
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
				m.Error("Hash doesn't match")
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

			err = m.MpcClient.Sign(ctx, nextPendingSignReq) // todo: add shared context to task manager
			m.Debug("Task-manager sent next sign request", logger.Field{"nextSignRequest", nextPendingSignReq})
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
				m.Error("Hash doesn't match")
				return hashMismatchErr
			}
			err = task.SetAddDelegatorTxSig(sig)
			if err != nil {
				return err
			}
			delete(m.pendingSignRequests, signReqId)
			m.Info("Mpc-manager: Cool! All signings for a stake task all done.")

			ids, err := m.Staker.IssueStakeTaskTxs(ctx, task)

			//err = doStake(task)
			if err != nil {
				m.Error("Failed to doStake",
					logger.Field{"error", err})
				return errors.WithStack(err)
			}
			m.Info("Mpc-manager: Cool! Success to add delegator!",
				logger.Field{"stakeTaske", task},
				logger.Field{"ids", ids})
		} else {
			m.Error("Hash doesn't match")
			return wrongRequestNumberErr
		}
	}
	return nil
}

func (m *Manager) watchStakeRequest(ctx context.Context) error {
	// Subscribe StakeRequestAdded event
	pubkeys, err := m.GetPubKeys(ctx, m.PubKeyHash.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	sinkAdded, err := m.WatchStakeRequestAdded(ctx, pubkeys)
	if err != nil {
		return errors.WithStack(err)
	}

	// Subscribe StakeRequestStarted event
	sinkStarted, err := m.WatchStakeRequestStarted(ctx, pubkeys)
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch StakeRequestAdded and StakeRequestStarted event
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-sinkAdded:
			m.WarnOnNotOk(ok, "Retrieve nothing from event channel of StakeRequestAdded")
			if ok {
				m.stakeRequestAddedEvt <- evt
			}
		case evt, ok := <-sinkStarted:
			m.WarnOnNotOk(ok, "Retrieve nothing from event channel of StakeRequestStarted")
			if ok {
				m.stakeRequestStartedEvt <- evt
			}
		}
	}
}

// todo: store this event info
func (m *Manager) onStakeRequestAdded(ctx context.Context, req *contract.MpcManagerStakeRequestAdded) error {
	ind, err := m.GetIndex(ctx, m.PubKeyHash.Hex(), req.PublicKey.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	tx, err := m.JoinRequest(ctx, m.Signer, req.RequestId, ind)
	if err != nil {
		m.Error("Failed to join stake request", []logger.Field{{"error", err}, {"tx", tx}}...)
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

func (m *Manager) onStakeRequestStarted(ctx context.Context, req *contract.MpcManagerStakeRequestStarted) error {
	m.removePendingJoin(req.RequestId)

	myInd, err := m.GetIndex(ctx, m.PubKeyHash.Hex(), req.PublicKey.Hex())
	if err != nil {
		return errors.WithStack(err)
	}

	var participating bool
	for _, ind := range req.ParticipantIndices {
		participating = participating || ind.Cmp(myInd) == 0
	}

	if !participating {
		// Not Participating, Ignore
		m.Info("Not participated to stake request", logger.Field{"stakeReqId", req.RequestId})
		return nil
	}

	nodeID, err := ids.ShortFromPrefixedString(req.NodeID, constants.NodeIDPrefix)

	if err != nil {
		return errors.WithStack(err)
	}

	pkHashHex := req.PublicKey.Hex()
	genPkInfo, err := m.LoadGeneratedPubKeyInfo(ctx, pkHashHex)
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

	nonce, err := m.NonceAt(ctx, *address, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	bl, _ := m.BalanceAt(ctx, *address, nil)
	m.Debug("$$$$$$$$$C Balance of C-Chain address before export", []logger.Field{
		{"address", *address},
		{"balance", bl.Uint64()}}...)

	baseFeeGwei := uint64(300) // TODO: It should be given by the contract

	nAVAXAmount := new(big.Int).Div(req.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !req.StartTime.IsUint64() || !req.EndTime.IsUint64() {
		return errors.New("invalid uint64")
	}
	task, err := NewStakeTask(m.NetworkContext, *pubkey, nonce, nodeID, nAVAXAmount.Uint64(), req.StartTime.Uint64(), req.EndTime.Uint64(), baseFeeGwei)
	if err != nil {
		return errors.WithStack(err)
	}
	taskId := req.Raw.TxHash.Hex()
	m.stakeTasks[taskId] = task
	hashBytes, err := task.ExportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}
	pariticipantKeys, err := m.GetPariticipantKeys(ctx, req.PublicKey.Hex(), req.ParticipantIndices)
	if err != nil {
		return errors.WithStack(err)
	}

	m.Debug("Task-manager got participant public keys", logger.Field{"participantPubKeys", pariticipantKeys})

	normalized, err := myCrypto.NormalizePubKeys(pariticipantKeys)
	m.Debug("Task-manager normalized participant public keys",
		logger.Field{"normalizedParticipantPubKeys", normalized})
	if err != nil {
		return errors.Wrapf(err, "failed to normalized participant public keys: %v", pariticipantKeys)
	}

	genPkInfo, err = m.LoadGeneratedPubKeyInfo(ctx, pkHashHex)
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
	err = m.MpcClient.Sign(ctx, request) // todo: add shared context to task manager
	if err != nil {
		return errors.WithStack(err)
	}
	m.pendingSignRequests[request.RequestId] = request
	return nil
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
