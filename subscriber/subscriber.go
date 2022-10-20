package subscriber

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

type Subscriber struct {
	ctx           context.Context
	logger        logger.Logger
	config        *Config
	client        *ethclient.Client
	subscription  ethereum.Subscription
	eventLogQueue Queue
}
type Queue interface {
	Enqueue(value interface{}) error
}

func NewSubscriber(ctx context.Context, logger logger.Logger, config *Config, eventLogQueue Queue) (*Subscriber, error) {
	return &Subscriber{
		ctx:           ctx,
		logger:        logger,
		config:        config,
		client:        nil,
		subscription:  nil,
		eventLogQueue: eventLogQueue,
	}, nil
}

func (s *Subscriber) Start() error {
	client, err := ethclient.Dial(s.config.EthWsURL)
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	s.client = client
	filter := ethereum.FilterQuery{
		Addresses: []common.Address{s.config.MpcManagerAddress},
	}

	eventLogs := make(chan types.Log, 1024)
	sub, err := client.SubscribeFilterLogs(s.ctx, filter, eventLogs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to contract events")
	}

	s.subscription = event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-eventLogs:
				s.eventLogQueue.Enqueue(log) // TODO: Handle failure. There should be only one publisher to this queue, so now it should be fine to ignore the error.
			case err := <-sub.Err():
				// TODO: Should we reconnect if this happens?
				return err
			case <-quit:
				return nil
			}
		}
	})
	return nil
}

func (s *Subscriber) Close() error {
	if s.subscription != nil {
		s.subscription.Unsubscribe()
		s.subscription = nil
	}
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
	return nil
}
