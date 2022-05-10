package contract

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
)

type Coordinator struct {
	*MpcCoordinator
	chainId *big.Int
	log     logger.Logger
}

// NewCoordinator creates a new singleton instance of MpcCoordinator, bound to a specific deployed coordinator.
func NewCoordinator(log logger.Logger, chainID *big.Int, contractAddress *common.Address, contractBackend bind.ContractBackend) (*Coordinator, error) {
	c, err := NewMpcCoordinator(*contractAddress, contractBackend)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new coordinator instance")
	}

	return &Coordinator{c, chainID, log}, nil
}

// todo: check receipt to see whether a group created successfully or not

func (c *Coordinator) CreateGroup_(txSenderPrivKey *ecdsa.PrivateKey, participantPubKeys [][]byte, threshold int64) (groupID []byte, err error) {
	signer, err := bind.NewKeyedTransactorWithChainID(txSenderPrivKey, c.chainId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create a transaction signer with private key %q.", txSenderPrivKey)
	}

	txn, err := c.CreateGroup(signer, participantPubKeys[:], big.NewInt(threshold))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create group.")
	}

	c.log.Debug("Sent a transaction to create group.",
		logger.Field{"participants", len(participantPubKeys)},
		logger.Field{"threshold", threshold},
		logger.Field{"txHashHex", txn.Hash().Hex()})

	return []byte(gofakeit.UUID()), nil // todo: return real created group id, maybe a ethclient is required, see deploy.go as a exmaple.
}

func (c *Coordinator) RequestKeygen_(txSenderPrivKey *ecdsa.PrivateKey, groupId []byte) error {
	signer, err := bind.NewKeyedTransactorWithChainID(txSenderPrivKey, c.chainId)
	if err != nil {
		return errors.Wrapf(err, "failed to create a transaction signer with private key %q.", txSenderPrivKey)
	}

	var groupId32 [32]byte
	copy(groupId32[:], groupId)

	txn, err := c.RequestKeygen(signer, groupId32)
	if err != nil {
		return errors.Wrapf(err, "failed to request keygen.")
	}
	c.log.Debug("Sent a transaction for keygen.", logger.Field{"txHashHex", txn.Hash().Hex()})
	return nil
}

func (c *Coordinator) RequestStake_(txSenderPrivKey *ecdsa.PrivateKey, publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) error {
	signer, err := bind.NewKeyedTransactorWithChainID(txSenderPrivKey, c.chainId)
	if err != nil {
		return errors.Wrapf(err, "failed to create a transaction signer with private key %q.", txSenderPrivKey)
	}

	txn, err := c.RequestStake(signer, publicKey, nodeID, amount, startTime, endTime)
	if err != nil {
		return errors.Wrapf(err, "failed to request stake.")
	}
	c.log.Debug("Sent a transaction for stake.", logger.Field{"txHashHex", txn.Hash().Hex()})
	return nil
}
