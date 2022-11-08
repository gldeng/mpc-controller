package syncer

import (
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
)

type Queue interface {
	Enqueue(value interface{}) error
}

type Syncer struct {
	services      *core.ServicePack
	eventLogQueue Queue
}

func NewSyncer(services *core.ServicePack, queue Queue) *Syncer {
	return &Syncer{services: services, eventLogQueue: queue}
}

func (s *Syncer) Start() error {
	ethClient := s.services.Config.CreateEthClient()

	instance, err := contract.NewMpcManagerFilterer(s.services.Config.MpcManagerAddress, ethClient)
	if err != nil {
		return err
	}
	it, err := instance.FilterParticipantAdded(nil, [][]byte{s.services.Config.MyPublicKey})
	if err != nil {
		return err
	}
	var groupIds [][32]byte
	for it.Next() {
		groupIds = append(groupIds, it.Event.GroupId)
		err = s.eventLogQueue.Enqueue(it.Event.Raw)
		if err != nil {
			return err
		}
	}

	it1, err := instance.FilterKeyGenerated(nil, groupIds)
	for it1.Next() {
		fmt.Printf("x\n")
		err = s.eventLogQueue.Enqueue(it1.Event.Raw)
		if err != nil {
			return err
		}
	}
	return nil
}
