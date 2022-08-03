package watcher

// Subscribe event:

// Process event: *events.KeygenRequestAdded
//
//type KeygenRequestAdded struct {
//	L    logger.L
//	GroupIDs  [][32]byte
//	Filterer  *contract.MpcManagerFilterer
//	P dispatcher.P
//	sub       event.Subscription
//	closeCh   chan struct{}
//}
//
//func (w *KeygenRequestAdded) Watch(ctx context.Context) error {
//	sink := make(chan *contract.MpcManagerKeygenRequestAdded)
//	sub, err := w.Filterer.WatchKeygenRequestAdded(nil, sink, w.GroupIDs)
//	if err != nil {
//		return errors.Wrapf(err, "failed to watch KeygenRequestAdded")
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
//				evtObj := dispatcher.NewEvtObj((*events.KeygenRequestAdded)(evt), nil)
//				w.P.Publish(ctx, evtObj)
//			case err := <-sub.Err():
//				w.L.ErrorOnError(err, "Got an error watching KeygenRequestAdded")
//			}
//		}
//	}()
//	return nil
//}
//
//func (w *KeygenRequestAdded) Close() {
//	close(w.closeCh)
//}
