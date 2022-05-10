package config

import (
	"crypto/ecdsa"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"

	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/goccy/go-yaml"
	"io/ioutil"
)

var _ Config = (*config)(nil)

type Config interface {
	IsDevMode() bool

	ControllerId() string
	ControllerKey() *ecdsa.PrivateKey
	ControllerSigner() *bind.TransactOpts

	MpcClient() core.MPCClient

	EthRpcClient() *ethclient.Client
	EthWsClient() *ethclient.Client

	CChainIssueClient() evm.Client
	PChainIssueClient() platformvm.Client

	CoordinatorAddress() *common.Address
	CoordinatorBoundInstance() *contract.MpcCoordinator
	CoordinatorBoundListener() *contract.MpcCoordinator

	NetworkContext() *core.NetworkContext
}

type config struct {
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

	mpcClient core.MPCClient

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

// todo: add config validator

func ParseConfig(filename string) Config {
	// Read config file
	cBytes, err := ioutil.ReadFile(filename)
	logger.FatalOnError(err, "Failed to read config file",
		logger.Field{"filename", filename},
		logger.Field{"error", err})

	// Unmarshal config content
	var c config
	err = yaml.Unmarshal(cBytes, &c)
	logger.FatalOnError(err, "Failed to read unmarshal config file",
		logger.Field{"filename", filename},
		logger.Field{"error", err})

	// Parse private key
	key, err := crypto.HexToECDSA(c.ControllerKey_)
	logger.FatalOnError(err, "Failed to parse secp256k1 private key",
		logger.Field{"key", c.ControllerKey_},
		logger.Field{"error", err})
	c.controllerKey = key

	// Convert chain ID
	chainIdBigInt := big.NewInt(c.ChainId)
	c.chainId = chainIdBigInt

	// Create controller transaction signer
	signer, err := bind.NewKeyedTransactorWithChainID(c.controllerKey, c.chainId)
	logger.FatalOnError(err, "Failed to create controller transaction signer",
		logger.Field{"key", c.ControllerKey_},
		logger.Field{"chainId", c.ChainId},
		logger.Field{"error", err})
	c.controllerSigner = signer

	// Create mpc-client
	mpcClient, err := core.NewMpcClient(c.MpcServerUrl)
	logger.FatalOnError(err, "Failed to create mpc-client",
		logger.Field{"url", c.MpcServerUrl},
		logger.Field{"error", err})
	c.mpcClient = mpcClient

	// Create eth rpc client
	ethRpcCli, err := ethclient.Dial(c.EthRpcUrl)
	logger.FatalOnError(err, "Failed to connect eth rpc client",
		logger.Field{"url", c.EthRpcUrl},
		logger.Field{"error", err})
	c.ethRpcClient = ethRpcCli

	// Create eth ws client
	ethWsCli, err := ethclient.Dial(c.EthWsUrl)
	logger.FatalOnError(err, "Failed to connect eth ws client",
		logger.Field{"url", c.EthWsUrl},
		logger.Field{"error", err})
	c.ethWsClient = ethWsCli

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
	c.coordinatorBoundInstance = coordBoundInst
	coordBoundListener, err := contract.NewMpcCoordinator(*c.coordinatorAddress, c.ethWsClient)
	c.coordinatorBoundListener = coordBoundListener

	// Convert C-Chain ID
	cchainID, err := ids.FromString(c.CChainId)
	logger.FatalOnError(err, "Failed to convert C-Chain ID",
		logger.Field{"cChainId", c.CChainId},
		logger.Field{"error", err})
	c.cChainId = &cchainID

	// Convert AVAX assetId ID
	assetId, err := ids.FromString(c.AvaxId)
	logger.FatalOnError(err, "Failed to convert AVAX assetId ID",
		logger.Field{"avaxId", c.AvaxId},
		logger.Field{"error", err})
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

	logger.Info("Config parsed successfully.")
	return &c
}

func (c *config) IsDevMode() bool {
	return c.EnableDevMode
}

func (c *config) ControllerId() string {
	return c.ControllerId_
}

func (c *config) ControllerKey() *ecdsa.PrivateKey {
	return c.controllerKey
}

func (c *config) ControllerSigner() *bind.TransactOpts {
	return c.controllerSigner
}

func (c *config) MpcClient() core.MPCClient {
	return c.mpcClient
}

func (c *config) EthRpcClient() *ethclient.Client {
	return c.ethRpcClient
}

func (c *config) EthWsClient() *ethclient.Client {
	return c.ethWsClient
}

func (c *config) CChainIssueClient() evm.Client {
	return c.cChainIssueClient
}

func (c *config) PChainIssueClient() platformvm.Client {
	return c.pChainIssueClient
}

func (c *config) CoordinatorAddress() *common.Address {
	return c.coordinatorAddress
}

func (c *config) CoordinatorBoundInstance() *contract.MpcCoordinator {
	return c.coordinatorBoundInstance
}

func (c *config) CoordinatorBoundListener() *contract.MpcCoordinator {
	return c.coordinatorBoundListener
}

func (c *config) NetworkContext() *core.NetworkContext {
	return c.networkContext
}
