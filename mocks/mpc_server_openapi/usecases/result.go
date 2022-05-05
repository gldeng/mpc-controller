package usecases

import (
	"context"
	"fmt"
	"github.com/swaggest/usecase"
)

func Result() usecase.IOInteractor {
	u := usecase.NewIOI(new(ResultInput), new(ResultOutput), func(ctx context.Context, input, output interface{}) error {
		var (
			in  = input.(*ResultInput)
			out = output.(*ResultOutput)
		)

		fmt.Printf("received result request, request id:%", in.RequestId)
		_ = out
		return nil
	})

	u.SetTitle("Query key or sign result")

	return u
}
