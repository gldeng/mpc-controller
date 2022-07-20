package backoff

import (
	"context"
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

func RetryFnConstant(ctx context.Context, maxRetries int, dur time.Duration, fn Fn) error {
	policy := ConstantPolicy(maxRetries, dur)
	return RetryFn(ctx, policy, fn)
}

func RetryFnConstant10Times(ctx context.Context, dur time.Duration, fn Fn) error {
	policy := ConstantPolicy(10, dur)
	return RetryFn(ctx, policy, fn)
}

func RetryFnConstantForever(ctx context.Context, dur time.Duration, fn Fn) error { // handy function
	policy := ConstantPolicy(0, dur)
	return RetryFn(ctx, policy, fn)
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

func RetryFnExponential(ctx context.Context, maxRetries int, minInterval, maxInterval time.Duration, fn Fn) error {
	policy := ExponentialPolicy(maxRetries, minInterval, maxInterval)
	return RetryFn(ctx, policy, fn)
}

func RetryFnExponential10Times(ctx context.Context, minInterval, maxInterval time.Duration, fn Fn) error {
	policy := ExponentialPolicy(10, minInterval, maxInterval)
	return RetryFn(ctx, policy, fn)
}

func RetryFnExponentialForever(ctx context.Context, minInterval, maxInterval time.Duration, fn Fn) error { // handy function
	policy := ExponentialPolicy(0, minInterval, maxInterval)
	return RetryFn(ctx, policy, fn)
}

// ----------

type Fn func() (retry bool, err error)

func RetryFn(ctx context.Context, policy backoff.Policy, fn Fn) error {
	b := policy.Start(ctx)
	var errStack error
	var startAt = time.Now()
	var retryNum = 0
	for backoff.Continue(b) {
		retry, err := fn()
		errStack = errors.WithStack(err)
		if !retry {
			break
		}
		retryNum++
	}
	retryDur := time.Now().Sub(startAt).Seconds()
	return errors.Wrapf(errStack, "exited after retrying %v times in %v seconds", retryNum, retryDur)
}
