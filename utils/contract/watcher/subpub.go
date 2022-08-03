package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/event"
)

type SubPub struct {
	Subscribe Subscribe
	Publish   Publish
}

type Subscribe func(ctx context.Context, closeCh chan struct{}, filterer interface{}) (sink chan interface{}, evt event.Subscription, err error)
type Publish func(ctx context.Context, publisher dispatcher.Publisher, evt interface{})
