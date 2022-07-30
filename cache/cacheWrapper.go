package cache

import (
	"context"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"sync"
)

type CacheWrapper struct {
	Dispatcher dispatcher.Dispatcher
	*Cache
}

func (c *CacheWrapper) Start(ctx context.Context) error {
	c.subscribe()
	<-ctx.Done()
	return nil
}

func (c *CacheWrapper) subscribe() {
	cache := Cache{
		RWMutex:                new(sync.RWMutex),
		GroupInfoMap:           make(map[string]events.GroupInfo),
		ParticipantInfoMap:     make(map[string]events.ParticipantInfo),
		GeneratedPubKeyInfoMap: make(map[string]events.GeneratedPubKeyInfo),
	}
	c.Cache = &cache

	c.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, c.Cache)
	c.Dispatcher.Subscribe(&events.ParticipantInfoStoredEvent{}, c.Cache)
	c.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, c.Cache)
}
