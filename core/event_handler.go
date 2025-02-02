package core

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	EvtParticipantAdded   = "ParticipantAdded"
	EvtKeygenRequestAdded = "KeygenRequestAdded"
	EvtKeyGenerated       = "KeyGenerated"
	EvtStakeRequestAdded  = "StakeRequestAdded"
	EvtRequestStarted     = "RequestStarted"
)

type LogEventHandler interface {
	Handle(ctx EventHandlerContext, log types.Log) ([]Task, error)
}

type EventHandlerContext interface {
	GetLogger() logger.Logger
	GetContract() *bind.BoundContract
	GetDb() Store
	GetEventID(event string) common.Hash
}

type EventHandlerContextFactory = func() EventHandlerContext
