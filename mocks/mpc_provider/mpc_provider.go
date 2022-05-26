package mpc_provider

import (
	"context"
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/token"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

const (
	MinimumToEnsureBalance = 1_000_000_000_000_000_000 //
)

type MpcProvider struct {
	log            logger.Logger
	chainId        *big.Int
	rpcClient      *ethclient.Client
	wsClient       *ethclient.Client
	RpcCoordinator *contract.MpcManager
	WsCoordinator  *contract.MpcManager
	privateKey     *ecdsa.PrivateKey
	txSigner       *bind.TransactOpts
}

func New(log logger.Logger,
	chainId *big.Int,
	coordinatorAddr *common.Address,
	privKey *ecdsa.PrivateKey,
	rpcClient *ethclient.Client,
	wsClient *ethclient.Client) *MpcProvider {

	rpcCoordinator, err := contract.NewMpcManager(*coordinatorAddr, rpcClient)
	log.FatalOnError(err, "Failed to create MpcManager bindings", logger.Field{"error", err})
	wsCoordinator, err := contract.NewMpcManager(*coordinatorAddr, wsClient)
	log.FatalOnError(err, "Failed to create MpcManager bindings", logger.Field{"error", err})

	signer, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	log.FatalOnError(err, "Failed to create transaction signer", logger.Field{"error", err})

	return &MpcProvider{
		log:            log,
		chainId:        chainId,
		rpcClient:      rpcClient,
		wsClient:       wsClient,
		RpcCoordinator: rpcCoordinator,
		WsCoordinator:  wsCoordinator,
		privateKey:     privKey,
		txSigner:       signer,
	}
}

// CreateGroup creates group with coordinator smart contract and return the created group id
// todo: return receipt?
func (m *MpcProvider) CreateGroup(participantPubKeys []*ecdsa.PublicKey, threshold int64) (string, error) {
	if len(participantPubKeys) < 3 {
		return "", errors.New("Require at least three participants to create a group")
	}

	if m.RpcCoordinator == nil || m.WsCoordinator == nil {
		return "", errors.New("Nil coordinators provided")
	}

	participants := crypto.MarshalPubkeys(participantPubKeys)
	var participants_ [][]byte
	for _, participant := range participants {
		participants_ = append(participants_, participant[1:])
	}

	_, err := m.RpcCoordinator.CreateGroup(m.txSigner, participants_, big.NewInt(threshold))
	if err != nil {
		return "", errors.WithStack(err)
	}

	groupId, err := m.waitForAllParticipantsAdded(participants_)
	if err != nil {
		return "", errors.WithStack(err)
	}

	err = m.ensureBalance(crypto.PubkeysToAddresses(participantPubKeys))
	if err != nil {
		return "", errors.WithStack(err)
	}

	logger.Info("Group created", logger.Field{"groupId", groupId})
	return groupId, nil
}

func (m *MpcProvider) RequestKeygen(groupIdHex string) (string, error) {
	groupId := common.FromHex(groupIdHex)

	type resultT struct {
		pubKeyHex string
		err       error
	}
	resultChan := make(chan resultT)
	go func() {
		logger.Debug("Staker started watch KeyGenerated event", logger.Field{"groupIdHex", groupIdHex})
		pubKeyHex, err := m.watchKeyGeneratedEvent(groupId)
		if err != nil {
			resultChan <- resultT{"", errors.WithStack(err)}
			return
		}
		resultChan <- resultT{pubKeyHex, nil}

	}()

	time.Sleep(time.Second * 2)
	var groupId32 [32]byte
	copy(groupId32[:], groupId)
	tx, err := m.RpcCoordinator.RequestKeygen(m.txSigner, groupId32)
	if err != nil {
		return "", errors.WithStack(err)
	}
	m.log.Info("Staker RequestKeygen sent.", logger.Field{"groupIdHex", groupIdHex})

	time.Sleep(time.Second * 5)
	rcp, err := m.rpcClient.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		m.log.Error("Got an error when query transaction receipt", logger.Field{"error", err})
		return "", errors.Wrap(err, "got an error when query transaction receipt")
	}

	if rcp.Status != 1 {
		m.log.Error("Transaction failed", logger.Field{"receipt", spew.Sdump(rcp)})
		return "", errors.Errorf("transaction failed, receipt: %s", spew.Sdump(rcp))
	}

	result := <-resultChan
	close(resultChan)

	if result.err != nil {
		return "", errors.WithStack(result.err)
	}
	m.log.Info("Staker received KeyGenerated event",
		logger.Field{"groupIdHex", groupIdHex},
		logger.Field{"pubkeyHex", result.pubKeyHex})

	return result.pubKeyHex, nil
}

func (m *MpcProvider) watchKeyGeneratedEvent(groupId []byte) (string, error) {
	var groupId32 [32]byte
	copy(groupId32[:], groupId)

	events := make(chan *contract.MpcManagerKeyGenerated)

	//var start = uint64(1)
	//opts := new(bind.WatchOpts) // todo: to get more clear on opts meaning.
	//opts.Start = &start
	sub, err := m.WsCoordinator.WatchKeyGenerated(nil, events, [][32]byte{groupId32})
	if err != nil {
		return "", errors.WithStack(err)
	}

	var listenErr error
	var publicKeyHex string

listen:
	for {
		select {
		case err := <-sub.Err():
			listenErr = err
			break listen
		case evt := <-events:
			publicKeyHex = common.Bytes2Hex(evt.PublicKey)
			break listen
		}
	}

	sub.Unsubscribe()

	return publicKeyHex, listenErr
}

func (m *MpcProvider) ensureBalance(participantAddrs []*common.Address) error {
	for _, addr := range participantAddrs {
		bal, err := m.rpcClient.BalanceAt(context.Background(), *addr, nil)
		if err != nil {
			return errors.WithStack(err)
		}
		if bal.Cmp(big.NewInt(MinimumToEnsureBalance)) < 0 {
			err = token.TransferInCChain(m.rpcClient, m.chainId, m.privateKey, addr, big.NewInt(MinimumToEnsureBalance))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

// todo: check whether the emit group id is fully corresponding to the given participant public keys.
func (m *MpcProvider) waitForAllParticipantsAdded(participantPubKeys [][]byte) (string, error) {
	events := make(chan *contract.MpcManagerParticipantAdded)

	var start = uint64(1)
	opts := new(bind.WatchOpts)
	opts.Start = &start
	sub, err := m.WsCoordinator.WatchParticipantAdded(opts, events, participantPubKeys)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var listenErr error
	var groupIDHex string

listen:
	for {
		select {
		case err := <-sub.Err():
			listenErr = err
			break listen
		case evt := <-events:
			groupIDHex = common.Bytes2Hex(evt.GroupId[:])
			break listen
		}
	}

	sub.Unsubscribe()

	return groupIDHex, listenErr
}
