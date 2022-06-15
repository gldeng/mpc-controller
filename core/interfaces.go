package core

import (
	"context"
)

type Keygener interface {
	Keygen(ctx context.Context, request *KeygenRequest) error
}

type Signer interface {
	Sign(ctx context.Context, request *SignRequest) error
}

type Resulter interface {
	Result(ctx context.Context, reqId string) (*Result, error)
}

type KeygenDoner interface {
	KeygenDone(ctx context.Context, request *KeygenRequest) (res *Result, err error)
}

type SignDoner interface {
	SignDone(ctx context.Context, request *SignRequest) (res *Result, err error)
}

type ResultDoner interface {
	ResultDone(ctx context.Context, reqId string) (res *Result, err error)
}
