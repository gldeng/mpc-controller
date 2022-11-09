package eventhandlercontext

import (
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	_ core.EventHandlerContext = (*EventHandlerContextImp)(nil)
)

type EventHandlerContextImp struct {
	Logger      logger.Logger
	MpcContract *bind.BoundContract
	DB          core.Store
	abi         *abi.ABI
}

func (c *EventHandlerContextImp) GetLogger() logger.Logger {
	return c.Logger
}

func (c *EventHandlerContextImp) GetContract() *bind.BoundContract {
	return c.MpcContract
}

func (c *EventHandlerContextImp) GetDb() core.Store {
	return c.DB
}

func (c *EventHandlerContextImp) GetEventID(event string) common.Hash {
	return c.abi.Events[event].ID
}

func NewEventHandlerContextImp(services *core.ServicePack) (*EventHandlerContextImp, error) {
	ethClient := services.Config.CreateEthClient()
	boundInstance, err := contract.BindMpcManagerCaller(services.Config.MpcManagerAddress, ethClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bound contract")
	}
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get abi")
	}
	return &EventHandlerContextImp{
		Logger:      services.Logger,
		MpcContract: boundInstance,
		DB:          services.Db,
		abi:         abi,
	}, nil
}
