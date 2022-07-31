package reconnector

import (
	"context"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

// Subscribe event:

// Publish event: *events.ContractFiltererCreatedEvent

type ContractFilterReconnector struct {
	logger.Logger

	Updater network.EthWsClientUpdater

	Publisher dispatcher.Publisher

	createdNo int
}

func (c *ContractFilterReconnector) Start(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				client, isUpdated, err := c.Updater.NewEthWsClient(ctx)
				if err != nil {
					c.Error("Failed to check check connectivity of EthWsClient", []logger.Field{{"error", err}}...)
					break
				}

				if c.createdNo == 0 {
					time.Sleep(5) // wait for the event subscriber to get ready.
					c.publishEvent(ctx, client)
					c.createdNo++
					break
				}

				if isUpdated {
					c.publishEvent(ctx, client)
					c.createdNo++
				}
			}
		}
	}()

	<-ctx.Done()

	return nil
}

func (c *ContractFilterReconnector) publishEvent(ctx context.Context, client *ethclient.Client) {
	newEvt := events.ContractFiltererCreatedEvent{
		Filterer: client,
	}
	evtObj := dispatcher.NewEvtObj(&newEvt, nil)
	c.Publisher.Publish(ctx, evtObj)
	c.Debug("Eth websocket client published.")
}
