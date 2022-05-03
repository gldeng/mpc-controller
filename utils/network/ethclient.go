package network

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"log"
	"sync"
)

var defaultClient *ethclient.Client

// DefaultURL is for Avalanche C-Chain, it's value can be set by external users.
var DefaultURL = "http://localhost:9650/ext/bc/C/rpc"

var once = new(sync.Once)

// NewEthClient return a new object for Ethereum-compatible client.
func NewEthClient(url string) *ethclient.Client {
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("%+v", errors.Wrapf(err, "got an error when Dial %q", url))
	}
	return client
}

// DefaultEthClient return a singleton object for Ethereum-compatible client.
func DefaultEthClient() *ethclient.Client {
	once.Do(func() {
		if defaultClient == nil {
			defaultClient = NewEthClient(DefaultURL)
		}
	})
	return defaultClient
}

// ----------websocket client---------
var defaultWsClient *ethclient.Client

// DefaultWsURL is for Avalanche C-Chain, it's value can be set by external users.
var DefaultWsURL = "ws://127.0.0.1:9650/ext/bc/C/ws"

var onceWs = new(sync.Once)

// NewWsEthClient return a new object for Ethereum-compatible websocket client.
func NewWsEthClient(url string) *ethclient.Client {
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("%+v", errors.Wrapf(err, "got an error when Dial %q", url))
	}
	return client
}

// DefaultWsEthClient return a singleton object for Ethereum-compatible websocket client.
func DefaultWsEthClient() *ethclient.Client {
	onceWs.Do(func() {
		if defaultWsClient == nil {
			defaultWsClient = NewWsEthClient(DefaultWsURL)
		}
	})
	return defaultWsClient
}
