package watcher

//import (
//	"context"
//	"github.com/avalido/mpc-controller/contract"
//	"github.com/avalido/mpc-controller/events"
//	"github.com/avalido/mpc-controller/logger"
//	"github.com/avalido/mpc-controller/utils/dispatcher"
//	"github.com/ethereum/go-ethereum/event"
//	"github.com/pkg/errors"
//)
//
//// Subscribe event:
//
//// Process event: *events.ParticipantAdded
//
//type ParticipantAdded struct {
//	L    logger.L
//	GroupIDs   [][]byte
//	Filterer  *contract.MpcManagerFilterer
//	P dispatcher.P
//	sub       event.Subscription
//	closeCh   chan struct{}
//}
//
//func (w *ParticipantAdded) Watch(ctx context.Context) error {
//	sink := make(chan *contract.MpcManagerParticipantAdded)
//	sub, err := w.Filterer.WatchParticipantAdded(nil, sink, w.GroupIDs)
//	if err != nil {
//		return errors.Wrapf(err, "failed to watch ParticipantAdded")
//	}
//	w.sub = sub
//	w.closeCh = make(chan struct{})
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				sub.Unsubscribe()
//				return
//			case <-w.closeCh:
//				sub.Unsubscribe()
//				return
//			case evt := <-sink:
//				evtObj := dispatcher.NewEvtObj((*events.ParticipantAdded)(evt), nil)
//				w.P.Publish(ctx, evtObj)
//			case err := <-sub.Err():
//				w.L.ErrorOnError(err, "Got an error watching ParticipantAdded")
//			}
//		}
//	}()
//	return nil
//}
//
//func (w *ParticipantAdded) Close() {
//	close(w.closeCh)
//}
