package usecases

import (
	"context"
	"fmt"
	"github.com/swaggest/usecase"
)

func Keygen() usecase.IOInteractor {
	u := usecase.NewIOI(new(KeygenInput), nil, func(ctx context.Context, input, output interface{}) error {
		var (
			in = input.(*KeygenInput)
		)

		fmt.Println("received key gen request", in)
		return nil
	})

	u.SetTitle("Generate key")

	return u
}
