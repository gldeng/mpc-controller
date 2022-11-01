package core

import (
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

var (
	_ EventHandlerContext = (*EventHandlerContextImp)(nil)
)

type LogEventHandler interface {
	Handle(ctx EventHandlerContext, log types.Log) ([]Task, error)
}

type EventHandlerContext interface {
	GetLogger() logger.Logger
	GetContract() *bind.BoundContract
	GetDb() storage.SlimDb
	GetEventID(event string) common.Hash
}

type EventHandlerContextFactory = func() EventHandlerContext

type EventHandlerContextImp struct {
	Logger      logger.Logger
	MpcContract *bind.BoundContract
	DB          storage.SlimDb
	abi         *abi.ABI
}

func (c *EventHandlerContextImp) GetLogger() logger.Logger {
	return c.Logger
}

func (c *EventHandlerContextImp) GetContract() *bind.BoundContract {
	return c.MpcContract
}

func (c *EventHandlerContextImp) GetDb() storage.SlimDb {
	return c.DB
}

func (c *EventHandlerContextImp) GetEventID(event string) common.Hash {
	return c.abi.Events[event].ID
}

func NewEventHandlerContextImp(services *ServicePack) (*EventHandlerContextImp, error) {
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
