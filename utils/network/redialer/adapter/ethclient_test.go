package adapter

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	myBackOff "github.com/avalido/mpc-controller/utils/backoff"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEthClientReDialer_GetClient(t *testing.T) {
	logger.DevMode = true
	reDialer := EthClientReDialer{
		Logger:        logger.Default(),
		EthURL:        "ws://127.0.0.1:9650/ext/bc/C/ws",
		BackOffPolicy: myBackOff.ExponentialPolicy(10, time.Second, time.Second*10),
	}

	_, cliCh, err := reDialer.GetClient(context.Background())
	// Start your server at ethWsURL
	require.Nil(t, err)
	fmt.Println("Client created")
	// Close your server at ethWsURL
	// Start your server at ethWsURL
	<-cliCh
	fmt.Println("Client recreated")
}
