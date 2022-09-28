package txtracker

import (
	"context"
	"time"
)

type TxToTrack struct {
	ID string
	Tx string
}

type TxTracker struct{}

func (t *TxTracker) TrackPChainTx() {}

func (t *TxTracker) TrackChainTx() {}

func (t *TxTracker) trackTx(ctx context.Context) {
	tk := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			// check tx status
		}
	}
}
