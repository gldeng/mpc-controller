package adapter

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/network/redialer"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lestrrat-go/backoff/v2"
	"github.com/pkg/errors"
)

type EthClientReDialer struct {
	Logger        logger.Logger
	EthURL        string
	BackOffPolicy backoff.Policy
}

func (d *EthClientReDialer) GetClient(ctx context.Context) (client redialer.Client, ReDialedClientCh chan redialer.Client, err error) {
	dial := func(ctx context.Context) (client redialer.Client, err error) {
		client, err = ethclient.Dial(d.EthURL)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to dial %v to create ethcient", d.EthURL)
		}
		return
	}
	isConnected := func(ctx context.Context, client redialer.Client) error {
		myClient := client.(*ethclient.Client)
		_, err := myClient.NetworkID(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to check connection of  %v with ethclient", d.EthURL)
		}
		return nil
	}

	logger.DevMode = true
	reDialer := &redialer.ReDialer{
		Logger:        d.Logger,
		Dial:          dial,
		IsConnected:   isConnected,
		BackOffPolicy: d.BackOffPolicy,
	}
	client, ReDialedClientCh, err = reDialer.GetClient(context.Background())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return
}
