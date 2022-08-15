package redialer

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	myBackOff "github.com/avalido/mpc-controller/utils/backoff"
	"github.com/lestrrat-go/backoff/v2"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

type IReDialer interface {
	GetClient(ctx context.Context) (client Client, ReDialedClientCh chan Client, err error)
}

type Client interface{}
type Dial func(ctx context.Context) (client Client, err error)
type IsConnected func(ctx context.Context, client Client) error // nil: no, non-nil: yes

type ReDialer struct {
	logger.Logger
	Dial             Dial
	IsConnected      IsConnected
	BackOffPolicy    backoff.Policy
	disconnected     uint32 // 0: no, 1: yes
	client           Client
	reDialedClientCh chan Client
	once             sync.Once
}

func (d *ReDialer) GetClient(ctx context.Context) (client Client, clientCh chan Client, err error) {
	d.once.Do(func() {
		client, err = d.dial(ctx)
		if err != nil {
			err = errors.Wrapf(err, "failed to dial at first time")
			return
		}

		if d.IsConnected(ctx, client); err != nil {
			err = errors.Wrapf(err, "first client is disoconnected")
			return
		}

		d.client = client
		d.reDialedClientCh = make(chan Client)
		go d.redial(ctx)
	})

	if err != nil {
		return
	}

	if atomic.LoadUint32(&d.disconnected) == 0 {
		return d.client, d.reDialedClientCh, nil
	}

	err = myBackOff.RetryFn(ctx, d.BackOffPolicy, func() (bool, error) {
		if atomic.LoadUint32(&d.disconnected) == 1 {
			return true, errors.Wrapf(err, "client is disconnected")
		}
		return false, nil
	})
	err = errors.Wrapf(err, "no valid client")
	return d.client, d.reDialedClientCh, err
}

func (d *ReDialer) redial(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := d.IsConnected(ctx, d.client); err == nil {
				break
			}

			atomic.StoreUint32(&d.disconnected, 1)
			d.Logger.Debug("Client is disconnected.")

			client, err := d.dial(ctx)
			if err != nil {
				d.Logger.ErrorOnError(err, "Failed to redial.")
				break
			}
			d.client = client
			atomic.StoreUint32(&d.disconnected, 0)
			d.Logger.Debug("Client is reconnected.")

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 60):
				return
			case d.reDialedClientCh <- client:
			}
		}
	}
}

func (d *ReDialer) dial(ctx context.Context) (client Client, err error) {
	err = myBackOff.RetryFn(ctx, d.BackOffPolicy, func() (bool, error) {
		if client, err = d.Dial(ctx); err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to dial")
	return
}
