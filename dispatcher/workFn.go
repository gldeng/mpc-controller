package dispatcher

//type Workshop interface {
//	AddTask(ctx context.Context, t *work.Task)
//}
//
//func workFnFromEventHandler(eh EventHandler) work.WorkFn {
//	return func(ctx context.Context, args interface{}) {
//		evtObj := args.(*EventObject)
//		eh.Do(ctx, evtObj)
//	}
//}
//
//func workFnFromEventHandlers(ehs []EventHandler) []work.WorkFn {
//	var workFns []work.WorkFn
//	for _, eh := range ehs {
//		workFns = append(workFns, workFnFromEventHandler(eh))
//	}
//
//	return workFns
//}
