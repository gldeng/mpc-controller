package core

import (
	"context"
	"github.com/avalido/mpc-controller/core/types"
)

// todo: Prometheus metrics

type MpcClient interface {
	Keygen(ctx context.Context, req *types.KeygenRequest) error
	Sign(ctx context.Context, req *types.SignRequest) error
	Result(ctx context.Context, reqID string) (*types.Result, error)
}
