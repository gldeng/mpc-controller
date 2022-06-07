package wrappers

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
)

type MpcManagerFilterWrapper struct {
	logger.Logger
	*contract.MpcManagerFilterer
}

// todo: deal with websocket network disconnection

func (m *MpcManagerFilterWrapper) WatchParticipantAdded(ctx context.Context, publicKey [][]byte) (<-chan *contract.MpcManagerParticipantAdded, error) {
	sink := make(chan *contract.MpcManagerParticipantAdded)

	err := backoff.RetryFnExponentialForever(m.Logger, ctx, func() error {
		sub, err := m.MpcManagerFilterer.WatchParticipantAdded(nil, sink, publicKey)
		if err != nil {
			m.Error("Failed to watch ParticipantAdded event", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					sub.Unsubscribe()
				case err := <-sub.Err():
					m.ErrorOnError(err, "Got an error during watching ParticipantAdded event", logger.Field{"error", err})
				}
			}
		}()

		return nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return sink, nil
}
