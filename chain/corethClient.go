package chain

import (
	"context"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

type CorethClientWrapper struct {
	Logger logger.Logger
	ethclient.Client
}

func (c *CorethClientWrapper) EstimateBaseFee(ctx context.Context) (baseFee *big.Int, err error) {
	backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second*1, time.Second*10, func() error {
		baseFee, err = c.Client.EstimateBaseFee(ctx)
		if err != nil {
			err = errors.Wrapf(err, "failed to request EstimateBaseFee with corethclient.Client")
			return err
		}
		return nil
	})
	return
}
