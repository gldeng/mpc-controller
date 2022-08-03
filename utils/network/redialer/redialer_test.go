package redialer

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	myBackOff "github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedialer_GetClient(t *testing.T) {
	ethWsURL := "ws://127.0.0.1:9650/ext/bc/C/ws"
	dial := func(ctx context.Context) (client Client, err error) {
		client, err = ethclient.Dial(ethWsURL)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to dial %v to create ethcient", ethWsURL)
		}
		return
	}
	isConnected := func(ctx context.Context, client Client) error {
		myClient := client.(*ethclient.Client)
		_, err := myClient.NetworkID(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to check connection of  %v with ethclient", ethWsURL)
		}
		return nil
	}

	logger.DevMode = true
	redialer := &ReDialer{
		Logger:        logger.Default(),
		Dial:          dial,
		IsConnected:   isConnected,
		BackOffPolicy: myBackOff.ExponentialPolicy(10, time.Second, time.Second*10),
	}
	_, cliCh, err := redialer.GetClient(context.Background())
	// Start your server at ethWsURL
	require.Nil(t, err)
	fmt.Println("Client created")
	// Close your server at ethWsURL
	// Start your server at ethWsURL
	<-cliCh
	fmt.Println("Client recreated")
}
