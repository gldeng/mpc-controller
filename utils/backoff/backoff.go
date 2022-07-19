package backoff

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/lestrrat-go/backoff/v2"
	"github.com/pkg/errors"
	"time"
)

func ConstantPolicy(maxRetries int, dur time.Duration) backoff.Policy {
	p := backoff.Constant(
		backoff.WithMaxRetries(maxRetries), // 0 for forever retry
		backoff.WithInterval(dur))
	return p
}

func RetryFnConstant(log logger.Logger, ctx context.Context, maxRetries int, dur time.Duration, fn func() error) error {
	policy := ConstantPolicy(maxRetries, dur)
	return RetryFn(log, ctx, policy, fn)
}

func RetryFnConstant10Times(log logger.Logger, ctx context.Context, dur time.Duration, fn func() error) error {
	policy := ConstantPolicy(10, dur)
	return RetryFn(log, ctx, policy, fn)
}

func RetryFnConstantForever(log logger.Logger, ctx context.Context, dur time.Duration, fn func() error) error { // handy function
	policy := ConstantPolicy(0, dur)
	return RetryFn(log, ctx, policy, fn)
}

// ----------

func ExponentialPolicy(maxRetries int, minInterval, maxInterval time.Duration) backoff.Policy {
	p := backoff.Exponential(
		backoff.WithMaxRetries(maxRetries), // 0 for forever retry
		backoff.WithMinInterval(minInterval),
		backoff.WithMaxInterval(maxInterval),
		backoff.WithMultiplier(1.5),
		backoff.WithJitterFactor(0.05),
	)
	return p
}

func RetryFnExponential(log logger.Logger, ctx context.Context, maxRetries int, minInterval, maxInterval time.Duration, fn func() error) error {
	policy := ExponentialPolicy(maxRetries, minInterval, maxInterval)
	return RetryFn(log, ctx, policy, fn)
}

func RetryFnExponential10Times(log logger.Logger, ctx context.Context, minInterval, maxInterval time.Duration, fn func() error) error {
	policy := ExponentialPolicy(10, minInterval, maxInterval)
	return RetryFn(log, ctx, policy, fn)
}

func RetryFnExponentialForever(log logger.Logger, ctx context.Context, minInterval, maxInterval time.Duration, fn func() error) error { // handy function
	policy := ExponentialPolicy(0, minInterval, maxInterval)
	return RetryFn(log, ctx, policy, fn)
}

// ----------

func RetryFn(log logger.Logger, ctx context.Context, policy backoff.Policy, fn func() error) error {
	b := policy.Start(ctx)
	var errStack error
	var startAt = time.Now()
	var retryNum = 0
	for backoff.Continue(b) {
		err := fn() // todo: tell RetryFn whether continue retry, "fn func() (bool, error)", or just "fn func() bool"
		if err == nil {
			return nil
		}
		errStack = errors.WithStack(err)
		//log.Debug("Retry", []logger.Field{
		//	{"error", err},
		//	{"retryNum", retryNum},
		//	{"retryAfter", time.Now().Sub(startAt).Seconds()}}...)
		//startAt = time.Now()
		retryNum++
	}
	retryDur := time.Now().Sub(startAt).Seconds()
	return errors.Wrapf(errStack, "exited after retrying %v times in %v seconds", retryNum, retryDur)
}
