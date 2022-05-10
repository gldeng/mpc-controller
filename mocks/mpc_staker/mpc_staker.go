package mpc_staker

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/token"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/juju/errors"
	pkgErrors "github.com/pkg/errors"
	"math/big"
	"time"
)

type MpcStaker struct {
	log      logger.Logger
	cChainId *big.Int

	cHttpClient      *ethclient.Client
	cWebsocketCLient *ethclient.Client

	cHttpCoordinator      *contract.Coordinator
	cWebsocketCoordinator *contract.Coordinator

	cPrivateKey *ecdsa.PrivateKey
}

// todo: allow given logger
// todo: use given logger

func New(log logger.Logger,
	cChainId *big.Int,
	coordinatorAddr *common.Address,
	cPrivateKey *ecdsa.PrivateKey,
	cHttpClient *ethclient.Client,
	cWsClient *ethclient.Client) *MpcStaker {

	cHttpCoord, err := contract.NewCoordinator(log, cChainId, coordinatorAddr, cHttpClient)
	log.FatalOnError(err, "Failed to create http coordinator", logger.Field{"error", err})

	cWsCoord, err := contract.NewCoordinator(log, cChainId, coordinatorAddr, cWsClient)
	log.FatalOnError(err, "Failed to create websocket coordinator", logger.Field{"error", err})

	return &MpcStaker{
		log:                   log,
		cChainId:              cChainId,
		cHttpClient:           cHttpClient,
		cWebsocketCLient:      cWsClient,
		cHttpCoordinator:      cHttpCoord,
		cWebsocketCoordinator: cWsCoord,
		cPrivateKey:           cPrivateKey,
	}
}

//func NewCoordinator(log logger.Logger, chainID *big.Int, contractAddress *common.Address, contractBackend bind.ContractBackend) (*Coordinator, error) {

// todo: watch StakeRequestAdded, and StakeRequestStarted
func (m *MpcStaker) RequestStakeAfterKeyAdded(groupIdHex string, nodeId string, stakeAmount *big.Int, stakeDays int) error {
	pubKeyHex, err := m.requestKeygen(groupIdHex)
	if err != nil {
		return pkgErrors.WithStack(err)
	}

	err = m.requestStake(pubKeyHex, nodeId, stakeAmount, stakeDays)
	if err != nil {
		return pkgErrors.WithStack(err)
	}

	m.log.Info("Staker RequestStakeAfterKeyAdded end.")
	return nil
}

func (m *MpcStaker) requestStake(pubKeyHex string, nodeId string, stakeAmount *big.Int, stakeDays int) error {
	pubKeyBytes := common.Hex2Bytes(pubKeyHex)

	pubKey, err := crypto.UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		return pkgErrors.WithStack(err)
	}
	account := ethCrypto.PubkeyToAddress(*pubKey)

	// todo: to adjust amount needed
	amountForTxFee := big.NewInt(1_000_000_000_000_000_000)
	times := big.NewInt(1_000_0)
	amountForTxFee = new(big.Int).Mul(amountForTxFee, times)

	err = m.ensureBalance(&account, new(big.Int).Add(stakeAmount, amountForTxFee))
	if err != nil {
		return pkgErrors.WithStack(err)
	}

	fiveMins := int64(5 * 60)
	stakeDaysInSeconds := int64(stakeDays * 24 * 60 * 60)
	startTime := time.Now().Unix() + fiveMins
	endTime := startTime + stakeDaysInSeconds

	err = m.cHttpCoordinator.RequestStake_(m.cPrivateKey, pubKeyBytes, nodeId, stakeAmount, big.NewInt(startTime), big.NewInt(endTime))
	if err != nil {
		return errors.Trace(err)
	}

	m.log.Info("Staker RequestStake sent",
		logger.Field{"pubKeyHex", pubKeyHex},
		logger.Field{"nodeId", nodeId},
		logger.Field{"stakeAmount", stakeAmount},
		logger.Field{"stakeDays", stakeDays})
	return nil
}

func (m *MpcStaker) ensureBalance(stakeAccountAddr *common.Address, transferAmount *big.Int) error {
	err := token.TransferInCChain(m.cHttpClient, m.cChainId, m.cPrivateKey, stakeAccountAddr, transferAmount)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (m *MpcStaker) requestKeygen(groupIdHex string) (string, error) {

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
			resultChan <- resultT{"", pkgErrors.WithStack(err)}
			return
		}
		resultChan <- resultT{pubKeyHex, nil}

	}()

	time.Sleep(time.Second * 2)
	err := m.cHttpCoordinator.RequestKeygen_(m.cPrivateKey, groupId)
	if err != nil {
		return "", pkgErrors.WithStack(err)
	}
	m.log.Info("Staker RequestKeygen sent.", logger.Field{"groupIdHex", groupIdHex})

	result := <-resultChan
	close(resultChan)

	if result.err != nil {
		return "", pkgErrors.WithStack(result.err)
	}
	m.log.Info("Staker received KeyGenerated event",
		logger.Field{"groupIdHex", groupIdHex},
		logger.Field{"pubkeyHex", result.pubKeyHex})

	return result.pubKeyHex, nil
}

func (m *MpcStaker) watchKeyGeneratedEvent(groupId []byte) (string, error) {
	var groupId32 [32]byte
	copy(groupId32[:], groupId)

	events := make(chan *contract.MpcCoordinatorKeyGenerated)

	//var start = uint64(1)
	//opts := new(bind.WatchOpts) // todo: to get more clear on opts meaning.
	//opts.Start = &start
	sub, err := m.cWebsocketCoordinator.WatchKeyGenerated(nil, events, [][32]byte{groupId32})
	if err != nil {
		return "", errors.Trace(err)
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
