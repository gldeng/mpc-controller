package watcher

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

type Subscriber interface {
	Subscribe(ctx context.Context) (sink chan types.Log, evt event.Subscription, err error)
	Process(ctx context.Context, log types.Log) error
}
