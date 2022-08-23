package network

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"time"
)

type EthClientGetter interface {
	GetEthWsClient(ctx context.Context) (*ethclient.Client, error)
}

type EthWsClientUpdater interface {
	NewEthWsClient(ctx context.Context) (c *ethclient.Client, isUpdated bool, err error)
}

type EthWsClientDialer interface {
	GetEthWsClient() (*ethclient.Client, error)
}

type EthClientDialerImpl struct {
	logger.Logger

	EthWsUrl    string
	EthWsClient *ethclient.Client
}

func (e *EthClientDialerImpl) GetEthWsClient(ctx context.Context) (cli *ethclient.Client, err error) {
	err = backoff.RetryFn(ctx, backoff.ExponentialPolicy(10, time.Millisecond*100, time.Second*10), func() (bool, error) {
		_, err := e.EthWsClient.NetworkID(ctx)
		if err == nil {
			return false, nil
		}

		newClient, err := ethclient.Dial(e.EthWsUrl)
		if err != nil {
			return true, errors.WithStack(err)
		}
		e.EthWsClient = newClient
		return false, nil
	})
	err = errors.Wrapf(err, "failed to create eth websocket client")
	cli = e.EthWsClient
	return
}

func (e *EthClientDialerImpl) NewEthWsClient(ctx context.Context) (c *ethclient.Client, isUpdated bool, err error) {
	err = backoff.RetryFn(ctx, backoff.ExponentialPolicy(10, time.Millisecond*100, time.Second*10), func() (bool, error) {
		_, err = e.EthWsClient.NetworkID(ctx)
		if err == nil {
			return false, nil
		}
		newClient, err := ethclient.Dial(e.EthWsUrl)
		if err != nil {
			return true, errors.WithStack(err)
		}
		isUpdated = true
		e.EthWsClient = newClient
		return false, nil
	})
	err = errors.Wrapf(err, "failed to create eth websocket client")
	c = e.EthWsClient
	return
}
