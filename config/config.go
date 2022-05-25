package config

import (
	"context"
	"crypto/ecdsa"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/goccy/go-yaml"
)

var _ Config = (*ConfigImpl)(nil)

type Config interface {
	IsDevMode() bool

	ControllerId() string
	ControllerKey() *ecdsa.PrivateKey
	ControllerSigner() *bind.TransactOpts

	MpcClient() core.MpcClient

	EthRpcClient() *ethclient.Client
	EthWsClient() *ethclient.Client

	CChainIssueClient() evm.Client
	PChainIssueClient() platformvm.Client

	CoordinatorAddress() *common.Address
	SetCoordinatorAddress(address string)
	CoordinatorBoundInstance() *contract.MpcCoordinator
	CoordinatorBoundListener() *contract.MpcCoordinator
	CoordinatorBoundListenerRebuild(log logger.Logger, ctx context.Context) (*ethclient.Client, *contract.MpcCoordinator, error)

	NetworkContext() *core.NetworkContext

	DatabasePath() string
}

type ConfigImpl struct {
	EnableDevMode bool `yaml:"enableDevMode"`

	ControllerId_ string `yaml:"controllerId"`

	ControllerKey_ string `yaml:"controllerKey"`

	CoordinatorAddress_ string `yaml:"coordinatorAddress"`

	MpcServerUrl string `yaml:"mpcServerUrl"`

	EthRpcUrl string `yaml:"ethRpcUrl"`
	EthWsUrl  string `yaml:"ethWsUrl"`

	CChainIssueUrl string `yaml:"cChainIssueUrl"`
	PChainIssueUrl string `yaml:"pChainIssueUrl"`

	//

	controllerKey    *ecdsa.PrivateKey
	controllerSigner *bind.TransactOpts

	mpcClient core.MpcClient

	ethRpcClient *ethclient.Client
	ethWsClient  *ethclient.Client

	cChainIssueClient evm.Client
	pChainIssueClient platformvm.Client

	coordinatorAddress       *common.Address
	coordinatorBoundInstance *contract.MpcCoordinator
	coordinatorBoundListener *contract.MpcCoordinator

	//
	ConfigNetwork
	networkContext *core.NetworkContext

	//
	ConfigDbBadger
}

type ConfigNetwork struct {
	NetworkId uint32 `yaml:"networkId"`

	ChainId int64 `yaml:"chainId"`

	CChainId string `yaml:"cChainId"`

	AvaxId string `yaml:"avaxId"`

	ImportFee  uint64 `yaml:"importFee"`
	GasPerByte uint64 `yaml:"gasPerByte"`
	GasPerSig  uint64 `yaml:"gasPerSig"`
	GasFixed   uint64 `yaml:"gasFixed"`

	//

	chainId  *big.Int
	cChainId *ids.ID
	avaxId   *ids.ID
}

type ConfigDbBadger struct {
	BadgerDbPath string `yaml:"badgerDbPath"`
}

func ParseConfigFromFile(filename string) *ConfigImpl {
	// Read ConfigImpl file
	cBytes, err := ioutil.ReadFile(filename)
	logger.FatalOnError(err, "Failed to read ConfigImpl file", logger.Field{"error", err})

	return ParseConfigFromStr(string(cBytes))
}

func ParseConfigFromStr(configYmlStr string) *ConfigImpl {
	// Unmarshal ConfigImpl content
	var c ConfigImpl
	err := yaml.Unmarshal([]byte(configYmlStr), &c)
	logger.FatalOnError(err, "Failed to unmarshal ConfigImpl content", logger.Field{"error", err})

	return &c
}

func InitConfig(log logger.Logger, c *ConfigImpl) Config {
	// Parse private key
	key, err := crypto.HexToECDSA(c.ControllerKey_)
	logger.FatalOnError(err, "Failed to parse secp256k1 private key", logger.Field{"error", err})
	c.controllerKey = key

	// Convert chain ID
	chainIdBigInt := big.NewInt(c.ChainId)
	c.chainId = chainIdBigInt

	// Create controller transaction signer
	signer, err := bind.NewKeyedTransactorWithChainID(c.controllerKey, c.chainId)
	logger.FatalOnError(err, "Failed to create controller transaction signer", logger.Field{"error", err})
	c.controllerSigner = signer

	// Create mpc-client
	mpcClient, err := core.NewMpcClient(log, c.MpcServerUrl)
	logger.FatalOnError(err, "Failed to create mpc-client", logger.Field{"error", err})
	c.mpcClient = mpcClient

	// Create eth rpc client
	ethRpcCli, err := ethclient.Dial(c.EthRpcUrl)
	logger.FatalOnError(err, "Failed to connect eth rpc client", logger.Field{"error", err})
	c.ethRpcClient = ethRpcCli

	// Create eth ws client
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	err = backoff.RetryFn(logger.Default(), ctx, backoff.ExponentialForever(), func() error {
		ethWsCli, err := ethclient.Dial(c.EthWsUrl)
		if err != nil {
			return err
		}
		c.ethWsClient = ethWsCli
		return nil
	})
	logger.FatalOnError(err, "Failed to connect eth ws client", logger.Field{"error", err})

	if c.ethWsClient == nil {
		logger.Fatal("Ethereum websocket client is nil")
	}

	// Create C-Chain issue client
	cChainIssueCli := evm.NewClient(c.CChainIssueUrl, "C")
	c.cChainIssueClient = cChainIssueCli

	// Create P-Chain issue client
	pChainIssueCli := platformvm.NewClient(c.PChainIssueUrl)
	c.pChainIssueClient = pChainIssueCli

	// Convert coordinator address
	coordinatorAddr := common.HexToAddress(c.CoordinatorAddress_)
	c.coordinatorAddress = &coordinatorAddr

	// Create coordinator bound instance and listener
	coordBoundInst, err := contract.NewMpcCoordinator(*c.coordinatorAddress, c.ethRpcClient)
	logger.FatalOnError(err, "Failed to create mpc-coordinator instance", logger.Field{"error", err})
	c.coordinatorBoundInstance = coordBoundInst
	coordBoundListener, err := contract.NewMpcCoordinator(*c.coordinatorAddress, c.ethWsClient)
	logger.FatalOnError(err, "Failed to create mpc-coordinator listener", logger.Field{"error", err})
	c.coordinatorBoundListener = coordBoundListener

	// Convert C-Chain ID
	cchainID, err := ids.FromString(c.CChainId)
	logger.FatalOnError(err, "Failed to convert C-Chain ID", logger.Field{"error", err})
	c.cChainId = &cchainID

	// Convert AVAX assetId ID
	assetId, err := ids.FromString(c.AvaxId)
	logger.FatalOnError(err, "Failed to convert AVAX assetId ID", logger.Field{"error", err})
	c.avaxId = &assetId

	networkCtx := core.NewNetworkContext(
		c.NetworkId,
		*c.cChainId,
		c.chainId,
		avax.Asset{
			ID: *c.avaxId,
		},
		c.ImportFee,
		c.GasPerByte,
		c.GasPerSig,
		c.GasFixed,
	)

	c.networkContext = &networkCtx
	return c
}

func (c *ConfigImpl) IsDevMode() bool {
	return c.EnableDevMode
}

func (c *ConfigImpl) ControllerId() string {
	return c.ControllerId_
}

func (c *ConfigImpl) ControllerKey() *ecdsa.PrivateKey {
	return c.controllerKey
}

func (c *ConfigImpl) ControllerSigner() *bind.TransactOpts {
	return c.controllerSigner
}

func (c *ConfigImpl) MpcClient() core.MpcClient {
	return c.mpcClient
}

func (c *ConfigImpl) EthRpcClient() *ethclient.Client {
	return c.ethRpcClient
}

func (c *ConfigImpl) EthWsClient() *ethclient.Client {
	return c.ethWsClient
}

func (c *ConfigImpl) CChainIssueClient() evm.Client {
	return c.cChainIssueClient
}

func (c *ConfigImpl) PChainIssueClient() platformvm.Client {
	return c.pChainIssueClient
}

func (c *ConfigImpl) CoordinatorAddress() *common.Address {
	return c.coordinatorAddress
}

func (c *ConfigImpl) SetCoordinatorAddress(address string) {
	c.CoordinatorAddress_ = address
}

func (c *ConfigImpl) CoordinatorBoundInstance() *contract.MpcCoordinator {
	return c.coordinatorBoundInstance
}

func (c *ConfigImpl) CoordinatorBoundListener() *contract.MpcCoordinator {
	return c.coordinatorBoundListener
}

func (c *ConfigImpl) CoordinatorBoundListenerRebuild(log logger.Logger, ctx context.Context) (*ethclient.Client, *contract.MpcCoordinator, error) {
	// Create eth ws client
	err := backoff.RetryFn(log, ctx, backoff.ExponentialForever(), func() error {
		ethWsCli, err := ethclient.Dial(c.EthWsUrl)
		if err != nil {
			return errors.Wrap(err, "failed to connect eth ws client")
		}
		c.ethWsClient = ethWsCli
		return nil
	})

	if err != nil {
		log.Error("Failed to connect eth ws client", logger.Field{"error", err})
		return nil, nil, errors.WithStack(err)
	}

	if c.ethWsClient == nil {
		logger.Error("Ethereum websocket client is nil")
		return nil, nil, errors.New("Ethereum websocket cient is nil")
	}

	// Create coordinator bound listener
	coordBoundListener, err := contract.NewMpcCoordinator(*c.coordinatorAddress, c.ethWsClient)
	if err != nil {
		log.Error("Failed to create mpc-coordinator listener", logger.Field{"error", err})
		return nil, nil, errors.Wrap(err, "failed to  create mpc-coordinator listener")
	}

	c.coordinatorBoundListener = coordBoundListener
	return c.ethWsClient, c.coordinatorBoundListener, nil
}

func (c *ConfigImpl) NetworkContext() *core.NetworkContext {
	return c.networkContext
}

func (c *ConfigImpl) DatabasePath() string {
	return c.BadgerDbPath
}
