package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
)

type Watchers struct {
	Logger   logger.Logger
	Args     []Arg
	watchers []*Watcher
}

func (ws *Watchers) Watch(ctx context.Context) error {
	for _, arg := range ws.Args {
		w := &Watcher{
			Logger: ws.Logger,
			Arg:    arg,
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
