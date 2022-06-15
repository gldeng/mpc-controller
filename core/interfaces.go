package core

import (
	"context"
)

type Keygen interface {
	Keygen(ctx context.Context, request *KeygenRequest) error
}

type Sign interface {
	Sign(ctx context.Context, request *SignRequest) error
}

type Resulter interface {
	Result(ctx context.Context, reqId string) (*Result, error)
}

type KeygenDone interface {
	KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error)
}

type SignDone interface {
	SignDone(ctx context.Context, request *SignRequest) (res *Result, err error)
}

type ResultDone interface {
	ResultDone(ctx context.Context, reqId string) (res *Result, err error)
}
