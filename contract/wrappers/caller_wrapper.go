package wrappers

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"math/big"
)

type MpcManagerCallerWrapper struct {
	logger.Logger
	*contract.MpcManagerCaller
}

func (m *MpcManagerCallerWrapper) GetGroup(ctx context.Context, groupId [32]byte) (Participants [][]byte, Threshold *big.Int, err error) {
	var p [][]byte
	var t *big.Int

	err = backoff.RetryFnExponentialForever(m.Logger, ctx, func() error {
		group, err := m.MpcManagerCaller.GetGroup(nil, groupId)
		if err != nil {
			m.Error("Failed to query group", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		p = group.Participants
		t = group.Threshold
		return nil
	})

	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return p, t, nil
}
