package mpc_staker

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/avalido/mpc-controller/utils/token"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/juju/errors"
	pkgErrors "github.com/pkg/errors"
	"math/big"
	"os"
	"time"
)

type MpcStaker struct {
	cChainId int64

	cHttpClient      *ethclient.Client
	cWebsocketCLient *ethclient.Client

	cHttpCoordinator      *contract.Coordinator
	cWebsocketCoordinator *contract.Coordinator

	cPrivateKey *ecdsa.PrivateKey
}

func New(cChainId int64, cPrivateKey, cCoordinatorAddress, cHttpUrl, cWebsocketUrl string) *MpcStaker {
	cHttpClient := network.NewEthClient(cHttpUrl)
	cWebsocketClient := network.NewWsEthClient(cWebsocketUrl)

	cCAddress := common.HexToAddress(cCoordinatorAddress)

	cHttpCoordinator, err := contract.NewCoordinator(cChainId, &cCAddress, cHttpClient)
	if err != nil {
		logger.Error("Failed to create http coordinator", logger.Field{"error", err})
		os.Exit(1)
	}
	cWebsocketCoordinator, err := contract.NewCoordinator(cChainId, &cCAddress, cWebsocketClient)
	if err != nil {
		logger.Error("Failed to create websocket coordinator", logger.Field{"error", err})
		os.Exit(1)
	}

	cPrivateKey_, err := ethCrypto.HexToECDSA(cPrivateKey)
	if err != nil {
		logger.Error("Failed to parse C-Chain private key", logger.Field{"privateKeyHex", cPrivateKey_})
		os.Exit(1)
	}

	return &MpcStaker{
		cChainId:              cChainId,
		cHttpClient:           cHttpClient,
		cWebsocketCLient:      cWebsocketClient,
		cHttpCoordinator:      cHttpCoordinator,
		cWebsocketCoordinator: cWebsocketCoordinator,
		cPrivateKey:           cPrivateKey_,
	}
}

// todo: watch StakeRequestAdded, and StakeRequestStarted
func (m *MpcStaker) RequestStakeAfterKeyAdded(groupIdHex string, nodeId string, stakeAmount int64, stakeDays int) error {
	pubKeyHex, err := m.requestKeygen(groupIdHex)
	if err != nil {
		return pkgErrors.WithStack(err)
	}

	err = m.requestStake(pubKeyHex, nodeId, stakeAmount, stakeDays)
	if err != nil {
		return pkgErrors.WithStack(err)
	}

	logger.Info("RequestStakeAfterKeyAdded end.")
	return nil
}

func (m *MpcStaker) requestStake(pubKeyHex string, nodeId string, stakeAmount int64, stakeDays int) error {
	pubKeyBytes := common.Hex2Bytes(pubKeyHex)

	pubKey, err := crypto.UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		return pkgErrors.WithStack(err)
	}
	account := ethCrypto.PubkeyToAddress(*pubKey)

	err = m.ensureBalance(&account, stakeAmount+1_000_000_000)
	if err != nil {
		return pkgErrors.WithStack(err)
	}

	fiveMins := int64(5 * 60)
	stakeDaysInSeconds := int64(stakeDays * 24 * 60 * 60)
	startTime := time.Now().Unix() + fiveMins
	endTime := startTime + stakeDaysInSeconds

	err = m.cHttpCoordinator.RequestStake_(m.cPrivateKey, pubKeyBytes, nodeId, big.NewInt(stakeAmount), big.NewInt(startTime), big.NewInt(endTime))
	if err != nil {
		return errors.Trace(err)
	}

	logger.Info("RequestStake sent",
		logger.Field{"pubKeyHex", pubKeyHex},
		logger.Field{"nodeId", nodeId},
		logger.Field{"stakeAmount", stakeAmount},
		logger.Field{"stakeDays", stakeDays})
	return nil
}

func (m *MpcStaker) ensureBalance(stakeAccountAddr *common.Address, transferAmount int64) error {
	err := token.TransferInCChain(m.cHttpClient, m.cChainId, m.cPrivateKey, stakeAccountAddr, transferAmount)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (m *MpcStaker) requestKeygen(groupIdHex string) (string, error) {
	groupId := common.FromHex(groupIdHex)
	err := m.cHttpCoordinator.RequestKeygen_(m.cPrivateKey, groupId)
	if err != nil {
		return "", pkgErrors.WithStack(err)
	}
	logger.Info("RequestKeygen sent.", logger.Field{"groupId", groupIdHex})

	publicKeyHex, err := m.watchKeyGeneratedEvent(groupId)
	if err != nil {
		return "", pkgErrors.WithStack(err)
	}

	logger.Info("KeyGenerated",
		logger.Field{"groupIdHex", groupIdHex},
		logger.Field{"pubkeyHex", publicKeyHex})

	return publicKeyHex, nil
}

func (m *MpcStaker) watchKeyGeneratedEvent(groupId []byte) (string, error) {
	var groupId32 [32]byte
	copy(groupId32[:], groupId)

	events := make(chan *contract.MpcCoordinatorKeyGenerated)

	var start = uint64(1)
	opts := new(bind.WatchOpts) // todo: to get more clear on opts meaning.
	opts.Start = &start
	sub, err := m.cWebsocketCoordinator.WatchKeyGenerated(opts, events, [][32]byte{groupId32})
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
