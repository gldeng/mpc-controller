package backoff

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/lestrrat-go/backoff/v2"
	"github.com/pkg/errors"
	"time"
)

func ExponentialForever() backoff.Policy {
	p := backoff.Exponential(
		backoff.WithMaxRetries(0),
		backoff.WithMinInterval(time.Millisecond*500),
		backoff.WithMaxInterval(time.Second*5),
		backoff.WithMultiplier(1.5),
		backoff.WithJitterFactor(0.05),
	)
	return p
}

func ConstantForever(dur time.Duration) backoff.Policy {
	p := backoff.Constant(
		backoff.WithMaxRetries(0),
		backoff.WithInterval(dur))
	return p
}

func RetryFnConstantForever(log logger.Logger, ctx context.Context, dur time.Duration, fn func() error) error {
	policy := ConstantForever(dur)
	return RetryFn(log, ctx, policy, fn)
}

func RetryFnExponentialForever(log logger.Logger, ctx context.Context, fn func() error) error {
	policy := ExponentialForever()
	return RetryFn(log, ctx, policy, fn)
}

func RetryFn(log logger.Logger, ctx context.Context, policy backoff.Policy, fn func() error) error {
	b := policy.Start(ctx)
	var lastErr error
	var lastRetryAt = time.Now()
	var retryNum = 1
	for backoff.Continue(b) {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		log.Debug("Retry", []logger.Field{
			{"error", err},
			{"retryNum", retryNum},
			{"retryAfter", time.Now().Sub(lastRetryAt).Seconds()}}...)
		lastRetryAt = time.Now()
		retryNum++
	}
	return errors.Wrapf(lastErr, "failed to retry")
}
