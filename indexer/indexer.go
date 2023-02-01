package indexer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/common"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"sync"
	"time"
)

type Indexer struct {
	services         *core.ServicePack
	closeOnce        sync.Once
	onCloseCtx       context.Context
	onCloseCtxCancel func()
}

func NewIndexer(services *core.ServicePack) *Indexer {
	onCloseCtx, cancel := context.WithCancel(context.Background())
	return &Indexer{
		services:         services,
		closeOnce:        sync.Once{},
		onCloseCtx:       onCloseCtx,
		onCloseCtxCancel: cancel,
	}
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
				i.scanDelegators(client)
				i.scanUTXOs(client)
				i.services.TxIndex.PurgeOlderThan(time.Now().Add(-720 * time.Hour)) // Purge older than 30 days
				interval := time.Now().Sub(nextRun)
				timer.Reset(interval)
			}
		}
	}()

	return nil
}

func (i *Indexer) Close() error {
	i.closeOnce.Do(func() {
		i.onCloseCtxCancel()
	})
	return nil
}

func (i *Indexer) scanDelegators(client platformvm.Client) error {
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

func (i *Indexer) scanUTXOs(client platformvm.Client) error {

	addresses, err := common.LoadAllAddresses(i.services.Db)
	if err != nil {
		return err
	}

	var utxosAll [][]byte
	startAddr := ids.ShortEmpty
	startUtxoID := ids.Empty

	for {
		var utxos [][]byte
		utxos, startAddr, startUtxoID, err = client.GetUTXOs(context.Background(), addresses.List(), 1024, startAddr, startUtxoID) // TODO: give proper context
		if err != nil {
			return err
		}
		if utxos == nil {
			break
		}
		utxosAll = append(utxosAll, utxos...)
	}

	var txIDs []ids.ID
	for _, bytes := range utxosAll {
		utxo := &avax.UTXO{}
		_, err := txs.Codec.Unmarshal(bytes, utxo)
		if err != nil {
			return err
		}
		txIDs = append(txIDs, utxo.TxID)
	}

	for _, txId := range txIDs {
		txBytes, err := client.GetTx(context.Background(), txId)
		if err != nil {
			return err
		}
		tx, err := txs.Parse(txs.Codec, txBytes)
		if err != nil {
			return err
		}
		uTx := tx.Unsigned.(*txs.ImportTx)
		reqHash := types.RequestHash{}
		copy(reqHash[:], uTx.Memo)
		err = i.services.TxIndex.SetTxByType(reqHash, "Import", txId)
		if err != nil {
			return err
		}
	}

	return nil
}
