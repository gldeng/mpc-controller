package core

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type LogEventHandler interface {
	Handle(ctx EventHandlerContext, log types.Log) ([]Task, error)
}

type EventHandlerContext interface {
	GetLogger() logger.Logger
	GetContract() *bind.BoundContract
	GetDb() storage.DB
	GetEventID(event string) common.Hash
}
