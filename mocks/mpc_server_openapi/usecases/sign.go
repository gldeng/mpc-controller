package usecases

import (
	"context"
	"fmt"
	"github.com/swaggest/usecase"
)

func Sign() usecase.IOInteractor {
	u := usecase.NewIOI(new(SignInput), nil, func(ctx context.Context, input, output interface{}) error {
		var (
			in = input.(*SignInput)
		)

		fmt.Println("received sign request", in)
		return nil
	})

	u.SetTitle("Sign digest in hex format")

	return u
}
