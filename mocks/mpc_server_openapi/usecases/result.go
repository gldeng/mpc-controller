package usecases

import (
	"context"
	"github.com/swaggest/usecase"
)

func Result() usecase.IOInteractor {
	u := usecase.NewIOI(new(ResultInput), new(ResultOutput), func(ctx context.Context, input, output interface{}) error {
		var (
			in  = input.(*ResultInput)
			out = output.(*ResultOutput)
		)

		lastKeygenReq := storer.GetKeygenRequestModel(in.RequestId)
		if lastKeygenReq != nil {
			out.RequestId = in.RequestId
			out.RequestType = string(lastKeygenReq.reqType)
			out.RequestStatus = string(lastKeygenReq.status)
			out.Result = lastKeygenReq.result
			return nil
		}

		lastSignReq := storer.GetSignRequestModel(in.RequestId)
		if lastSignReq != nil {
			out.RequestId = in.RequestId
			out.RequestType = string(lastSignReq.reqType)
			out.RequestStatus = string(lastSignReq.status)
			out.Result = lastSignReq.result
			return nil
		}
		return nil
	})

	u.SetTitle("Query key or sign result")

	return u
}
