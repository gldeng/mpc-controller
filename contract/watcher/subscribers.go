package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/contract/watcher"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

const (
	EvtParticipantAdded   = "ParticipantAdded"
	EvtKeygenRequestAdded = "KeygenRequestAdded"
	EvtKeyGenerated       = "KeyGenerated"
	EvtStakeRequestAdded  = "StakeRequestAdded"
	EvtRequestStarted     = "RequestStarted"
)

type EvtName string

func NewSubscriber(evtName string, queries []interface{}) watcher.Subscriber {
	switch evtName {
	case EvtParticipantAdded:
	case EvtKeygenRequestAdded:
	case EvtKeyGenerated:
	case EvtStakeRequestAdded:
	case EvtRequestStarted:
	}
	return nil
}

func ParticipantAddedQueries(pubKeys [][]byte) []interface{} {
	var queries []interface{}
	for _, pubKey := range pubKeys {
		queries = append(queries, pubKey)
	}
	return queries
}

// Watch event: *contract.MpcManagerParticipantAdded
// Process event: *events.ParticipantAdded

type ParticipantAdded struct {
	PubKeys [][]byte
	L       logger.Logger
	P       dispatcher.Publisher
	B       bind.BoundContract
}

func (s *ParticipantAdded) Subscribe(ctx context.Context) (logs chan types.Log, sub event.Subscription, err error) {
	var queries []interface{}
	for _, pubKey := range s.PubKeys {
		queries = append(queries, pubKey)
	}
	return s.B.WatchLogs(nil, EvtParticipantAdded, queries)
}

func (s *ParticipantAdded) Process(ctx context.Context, log types.Log) error {
	myEvent := new(contract.MpcManagerParticipantAdded)
	if err := s.B.UnpackLog(myEvent, EvtParticipantAdded, log); err != nil {
		return err
	}
	myEvent.Raw = log

	myEvtObj := dispatcher.NewEvtObj((*events.ParticipantAdded)(myEvent), nil)
	s.P.Publish(ctx, myEvtObj)
	return nil
}

// Watch event: *contract.MpcManagerKeygenRequestAdded
// Process event: *events.KeygenRequestAdded

type KeygenRequestAdded struct {
	GroupIDs [][32]byte
	L        logger.Logger
	P        dispatcher.Publisher
	B        bind.BoundContract
}

func (s *KeygenRequestAdded) Subscribe(ctx context.Context) (logs chan types.Log, sub event.Subscription, err error) {
	var queries []interface{}
	for _, groupID := range s.GroupIDs {
		queries = append(queries, groupID)
	}
	return s.B.WatchLogs(nil, EvtKeygenRequestAdded, queries)
}

func (s *KeygenRequestAdded) Process(ctx context.Context, log types.Log) error {
	myEvent := new(contract.MpcManagerKeygenRequestAdded)
	if err := s.B.UnpackLog(myEvent, EvtKeygenRequestAdded, log); err != nil {
		return err
	}
	myEvent.Raw = log

	myEvtObj := dispatcher.NewEvtObj((*events.KeygenRequestAdded)(myEvent), nil)
	s.P.Publish(ctx, myEvtObj)
	return nil
}
