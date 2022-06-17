package contract

import (
	"context"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"time"
)

// Accept event:

// Emit event: *events.ContractFiltererCreatedEvent

type ContractFilterReconnector struct {
	logger.Logger

	EthWsUrl    string
	EthWsClient *ethclient.Client

	Publisher dispatcher.Publisher
}

func (c *ContractFilterReconnector) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for {
					if c.isDisconnected(ctx) {
						select {
						case <-ctx.Done():
							return
						default:
							err := c.reConnect(ctx)
							if err == nil {
								c.Error("Failed to reconnect websocket", []logger.Field{
									{"error", err},
									{"url", c.EthWsUrl}}...)
								break
							}

							select {
							case <-ctx.Done():
								return
							default:
								c.publishReconnectEvent(ctx)
							}
						}
					}
				}
			}
		}

	}()
	return nil
}

func (c *ContractFilterReconnector) isDisconnected(ctx context.Context) bool {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return false
		case <-ticker.C:
			_, err := c.EthWsClient.NetworkID(ctx)
			if err != nil {
				return true
			}
		}
	}
}

func (c *ContractFilterReconnector) reConnect(ctx context.Context) error {
	err := backoff.RetryFn(c.Logger, ctx, backoff.ExponentialForever(), func() error {
		client, err := ethclient.Dial(c.EthWsUrl)
		if err != nil {
			return errors.WithStack(err)
		}
		c.EthWsClient = client
		return nil
	})
	return errors.WithStack(err)
}

func (c *ContractFilterReconnector) publishReconnectEvent(ctx context.Context) {
	newEvt := events.ContractFiltererCreatedEvent{
		Filterer: c.EthWsClient,
	}

	c.Publisher.Publish(ctx, dispatcher.NewRootEventObject("ContractFilterReconnector", &newEvt, ctx))
}
