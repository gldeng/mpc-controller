package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/dispatcher"
)

type Watchers struct {
	Logger    logger.Logger
	SubPubS   []SubPub
	Filterer  interface{}
	Publisher dispatcher.Publisher
	watchers  []*Watcher
}

func (ws *Watchers) Watch(ctx context.Context) error {
	for _, subPub := range ws.SubPubS {
		w := &Watcher{
			Logger:    ws.Logger,
			Subscribe: subPub.Subscribe,
			Publish:   subPub.Publish,
			Filterer:  ws.Filterer,
			Publisher: ws.Publisher,
		}
		ws.watchers = append(ws.watchers, w)
	}
	for _, w := range ws.watchers {
		if err := w.Watch(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (ws *Watchers) Close() {
	for _, w := range ws.watchers {
		w.Close()
	}
}
