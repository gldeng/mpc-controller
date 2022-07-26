package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/utils/work"
)

func workFnFromEventHandler(eh EventHandler) work.WorkFn {
	return func(ctx context.Context, args interface{}) {
		evtObj := args.(*EventObject)
		eh.Do(ctx, evtObj)
	}
}

func workFnFromEventHandlers(ehs []EventHandler) []work.WorkFn {
	var workFns []work.WorkFn
	for _, eh := range ehs {
		workFns = append(workFns, workFnFromEventHandler(eh))
	}

	return workFns
}
