package caller

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	bind2 "github.com/avalido/mpc-controller/utils/contract/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var _ Caller = (*MyCaller)(nil)

const (
	MethodGetGroup                 Method = "getGroup"
	MethodGetGroupIdByKey          Method = "getGroupIdByKey"
	MethodPrincipalTreasuryAddress Method = "principalTreasuryAddress"
	MethodRewardTreasuryAddress    Method = "rewardTreasuryAddress"
)

type Method string

type Caller interface {
	Call(ctx context.Context, fn CallFn) error
	GetGroup(ctx context.Context, opts *bind.CallOpts, groupId [32]byte) ([][]byte, error)
	GetGroupIdByKey(ctx context.Context, opts *bind.CallOpts, publicKey []byte) ([32]byte, error)
	PrincipalTreasuryAddress(ctx context.Context, opts *bind.CallOpts) (common.Address, error)
	RewardTreasuryAddress(ctx context.Context, opts *bind.CallOpts) (common.Address, error)
}

type CallFn func() (err error, retry bool)

type MyCaller struct {
	Logger         logger.Logger
	ContractAddr   common.Address
	ContractCaller bind.ContractCaller
	boundCaller    bind2.BoundCaller
}

func (t *MyCaller) Init(ctx context.Context) {
	boundCaller, err := contract.BindMpcManagerCaller(t.ContractAddr, t.ContractCaller)
	t.Logger.FatalOnError(err, "Failed to bind MpcManager caller")
	t.boundCaller = boundCaller
}

func (c *MyCaller) GetGroup(ctx context.Context, opts *bind.CallOpts, groupId [32]byte) ([][]byte, error) {
	var group [][]byte
	fn := func() (err error, retry bool) {
		var out []interface{}
		err = c.boundCaller.Call(opts, &out, string(MethodGetGroup), groupId)
		if err != nil {
			return errors.WithStack(err), true
		}
		out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)
		group = out0
		return nil, false
	}
	return group, errors.WithStack(c.Call(ctx, fn))
}

func (c *MyCaller) GetGroupIdByKey(ctx context.Context, opts *bind.CallOpts, publicKey []byte) ([32]byte, error) {
	var groupId [32]byte
	fn := func() (err error, retry bool) {
		var out []interface{}
		err = c.boundCaller.Call(opts, &out, string(MethodGetGroupIdByKey), publicKey)
		if err != nil {
			return errors.WithStack(err), true
		}
		out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
		groupId = out0
		return nil, false
	}
	return groupId, errors.WithStack(c.Call(ctx, fn))
}

func (c *MyCaller) PrincipalTreasuryAddress(ctx context.Context, opts *bind.CallOpts) (common.Address, error) {
	return c.address(ctx, opts, MethodPrincipalTreasuryAddress)
}

func (c *MyCaller) RewardTreasuryAddress(ctx context.Context, opts *bind.CallOpts) (common.Address, error) {
	return c.address(ctx, opts, MethodRewardTreasuryAddress)
}

func (c *MyCaller) Call(ctx context.Context, fn CallFn) error {
	err := backoff.RetryRetryFnForever(ctx, func() (retry bool, err error) {
		err, retry = fn()
		if retry {
			return true, errors.WithStack(err)
		}
		return false, errors.WithStack(err)
	})

	return errors.WithStack(err)
}

func (c *MyCaller) address(ctx context.Context, opts *bind.CallOpts, address Method) (common.Address, error) {
	var addr common.Address
	fn := func() (err error, retry bool) {
		var out []interface{}
		err = c.boundCaller.Call(opts, &out, string(address))
		if err != nil {
			return errors.WithStack(err), true
		}
		out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
		addr = out0
		return nil, false
	}
	return addr, errors.WithStack(c.Call(ctx, fn))
}
