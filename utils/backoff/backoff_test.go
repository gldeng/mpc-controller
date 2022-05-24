package backoff

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/lestrrat-go/backoff/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestExponentialForever(t *testing.T) {
	p := ExponentialForever()

	flakyFunc := func() error {
		return errors.New("failed to dial ...")
	}

	// It will stop retrying due to specified context timeout limit
	//ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	//defer cancel()

	// It will keep retying due to with no context timeout limit
	ctx := context.Background()

	retryFunc := func() error {
		b := p.Start(ctx)
		var lastErr error
		var lastRetryAt = time.Now()
		var retryNum = 1
		for backoff.Continue(b) {
			err := flakyFunc()
			if err == nil {
				return nil
			}
			lastErr = err
			fmt.Printf("Failed to get value, error: %v, retried after: %v seconds, %v times \n",
				err, time.Now().Sub(lastRetryAt).Seconds(), retryNum)
			lastRetryAt = time.Now()
			retryNum++
		}
		return errors.Wrapf(lastErr, "failed to get value")
	}

	err := retryFunc()
	require.NotNil(t, err)
}

func TestRetryFn(t *testing.T) {
	p := ExponentialForever()

	type scenario struct {
		name    string
		fn      func() error
		wantErr bool
	}

	flakyFuncErr := func() error {
		return errors.New("failed to dial ...")
	}

	flakyFuncNil := func() error {
		return nil
	}

	scenarios := []scenario{
		{"Fn returns error", flakyFuncErr, true},
		{"Fn returns nil", flakyFuncNil, false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := RetryFn(logger.Default(), ctx, p, scenario.fn)
			assert.True(t, err != nil == scenario.wantErr)
		})
	}
}
