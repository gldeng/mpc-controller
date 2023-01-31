package indexer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/common"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"time"
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
		nextRun := time.Now().Add(60 * time.Minute)

		timer := time.NewTimer(1 * time.Minute) // First run after 1 minute
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				i.runCurrentValidators(client)
				interval := time.Now().Sub(nextRun)
				timer.Reset(interval)
			}
		}
	}()

	return nil
}

func (i *Indexer) runCurrentValidators(client platformvm.Client) error {
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
		if i.services.TxIndex.IsKnownTx(txId) {
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
		uTx := tx.Unsigned.(*txs.AddDelegatorTx)
		reqHash := types.RequestHash{}
		copy(reqHash[:], uTx.Memo)
		err = i.services.TxIndex.SetTxByType(reqHash, "AddDelegator", txId)
		if err != nil {
			return err
		}
	}
	return nil
}
