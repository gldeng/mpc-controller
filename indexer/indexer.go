package indexer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/common"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
)

type Indexer struct {
	services *core.ServicePack
}

func NewIndexer(services *core.ServicePack) *Indexer {
	return &Indexer{services: services}
}

func (i *Indexer) Start() error {
	client := i.services.Config.CreatePClient()

	go func() {

	}()

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
		err = s.eventLogQueue.Enqueue(it1.Event.Raw)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Indexer) runCurrentValidators(client platformvm.Client, txIndex core.TxIndex) error {
	validators, err := client.GetCurrentValidators(context.Background(), ids.Empty, nil) // TODO: give proper context
	if err != nil {
		return err
	}

	addresses, err := common.LoadAllAddresses(i.services.Db)
	if err != nil {
		return err
	}
	var txIds []ids.ID
	for _, validator := range validators {
		for _, delegator := range validator.Delegators {
			for _, address := range delegator.RewardOwner.Addresses {
				if addresses.Contains(address) {
					txIds = append(txIds, delegator.TxID)
				}
			}
		}
	}
	for _, txId := range txIds {
		if txIndex.IsKnownTx(txId) {
			continue
		}
		txBytes, err := client.GetTx(context.Background(), txId)
		if err != nil {
			return err
		}
		tx, err := txs.Parse(txs.Codec, txBytes)
		if err != nil {
			return err
		}
	}
	return nil
}
