package reconnector

import (
	"context"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/network"
	"time"
)

// Accept event:

// Emit event: *events.ContractFiltererCreatedEvent

type ContractFilterReconnector struct {
	logger.Logger

	Updater network.EthWsClientUpdater

	Publisher dispatcher.Publisher
}

func (c *ContractFilterReconnector) Start(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
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
				if isUpdated {
					newEvt := events.ContractFiltererCreatedEvent{
						Filterer: client,
					}
					c.Publisher.Publish(ctx, dispatcher.NewRootEventObject("ContractFilterReconnector", &newEvt, ctx))
				}
			}
		}
	}()
	return nil
}
