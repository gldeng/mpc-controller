package watcher

import (
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	myBind "github.com/avalido/mpc-controller/utils/contract/bind"
	"github.com/avalido/mpc-controller/utils/contract/watcher"

	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

const (
	EvtParticipantAdded   EvtName = "ParticipantAdded"
	EvtKeygenRequestAdded EvtName = "KeygenRequestAdded"
	EvtKeyGenerated       EvtName = "KeyGenerated"
	EvtStakeRequestAdded  EvtName = "StakeRequestAdded"
	EvtRequestStarted     EvtName = "RequestStarted"
)

type EvtName string

type Factory struct {
	Logger        logger.Logger
	Publisher     dispatcher.Publisher
	BoundFilterer myBind.BoundFilterer
}

func (f *Factory) NewWatcher(process watcher.Process, opts *bind.WatchOpts, name EvtName, queries ...[]interface{}) (*watcher.Watcher, error) {
	var (
		logs   chan types.Log
		sub    event.Subscription
		err    error
		unpack func(log types.Log) (interface{}, error)
	)
	switch name {
	case EvtParticipantAdded:
		logs, sub, err = f.BoundFilterer.WatchLogs(opts, string(EvtParticipantAdded), queries...)
		unpack = func(log types.Log) (interface{}, error) {
			myEvent := new(contract.MpcManagerParticipantAdded)
			if err := f.BoundFilterer.UnpackLog(myEvent, string(EvtParticipantAdded), log); err != nil {
				return nil, errors.Wrapf(err, fmt.Sprintf("failed to unpack %v log", EvtParticipantAdded))
			}
			myEvent.Raw = log
			return myEvent, nil
		}
	case EvtKeygenRequestAdded:
		logs, sub, err = f.BoundFilterer.WatchLogs(opts, string(EvtKeygenRequestAdded), queries...)
		unpack = func(log types.Log) (interface{}, error) {
			myEvent := new(contract.MpcManagerKeygenRequestAdded)
			if err := f.BoundFilterer.UnpackLog(myEvent, string(EvtKeygenRequestAdded), log); err != nil {
				return nil, errors.Wrapf(err, fmt.Sprintf("failed to unpack %v log", EvtKeygenRequestAdded))
			}
			myEvent.Raw = log
			return myEvent, nil
		}
	case EvtKeyGenerated:
		logs, sub, err = f.BoundFilterer.WatchLogs(opts, string(EvtKeyGenerated), queries...)
		unpack = func(log types.Log) (interface{}, error) {
			myEvent := new(contract.MpcManagerKeyGenerated)
			if err := f.BoundFilterer.UnpackLog(myEvent, string(EvtKeyGenerated), log); err != nil {
				return nil, errors.Wrapf(err, fmt.Sprintf("failed to unpack %v log", EvtKeyGenerated))
			}
			myEvent.Raw = log
			return myEvent, nil
		}
	case EvtStakeRequestAdded:
		logs, sub, err = f.BoundFilterer.WatchLogs(opts, string(EvtStakeRequestAdded), queries...)
		unpack = func(log types.Log) (interface{}, error) {
			myEvent := new(contract.MpcManagerStakeRequestAdded)
			if err := f.BoundFilterer.UnpackLog(myEvent, string(EvtStakeRequestAdded), log); err != nil {
				return nil, errors.Wrapf(err, fmt.Sprintf("failed to unpack %v log", EvtStakeRequestAdded))
			}
			myEvent.Raw = log
			return myEvent, nil
		}
	case EvtRequestStarted:
		logs, sub, err = f.BoundFilterer.WatchLogs(opts, string(EvtRequestStarted), queries...)
		unpack = func(log types.Log) (interface{}, error) {
			myEvent := new(contract.MpcManagerRequestStarted)
			if err := f.BoundFilterer.UnpackLog(myEvent, string(EvtRequestStarted), log); err != nil {
				return nil, errors.Wrapf(err, fmt.Sprintf("failed to unpack %v log", EvtRequestStarted))
			}
			myEvent.Raw = log
			return myEvent, nil
		}
	}

	myWatcher := &watcher.Watcher{
		Logger: f.Logger,
		Arg:    watcher.Arg{Logs: logs, Sub: sub, Unpack: unpack, Process: process},
	}

	return myWatcher, errors.Wrapf(err, "failed to create watcher")
}
