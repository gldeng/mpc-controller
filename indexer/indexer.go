package indexer

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/common"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/pkg/errors"
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

		timer := time.NewTimer(1 * time.Minute) // First run after 1 minute
		defer timer.Stop()

		for {
			select {
			case <-i.onCloseCtx.Done():
				return
			case <-timer.C:
				nextRun := time.Now().Add(60 * time.Minute)
				err := i.scanDelegators(client)
				if err != nil {
					i.services.Logger.Error(err.Error())
				} else {
					i.services.Logger.Debug("finished scanning delegators")
				}
				err = i.scanUTXOs(client)
				if err != nil {
					i.services.Logger.Error(err.Error())
				} else {
					i.services.Logger.Debug("finished scanning utxos")
				}
				i.services.TxIndex.PurgeOlderThan(time.Now().Add(-720 * time.Hour)) // Purge older than 30 days
				interval := nextRun.Sub(time.Now())
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
	i.services.Logger.Debug("scanning delegators")
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
	i.services.Logger.Debug("scanning utxos")
	addresses, err := common.LoadAllAddresses(i.services.Db)
	if err != nil {
		return err
	}

	var utxosAll [][]byte
	nextStartAddr := ids.ShortEmpty
	nextStartUtxoID := ids.Empty

	runCount := 1
	maxEntries := uint32(1024)
	for {
		var utxos [][]byte
		utxos, endAddr, endUtxoID, err := client.GetUTXOs(context.Background(), addresses.List(), maxEntries, nextStartAddr, nextStartUtxoID) // TODO: give proper context
		if err != nil {
			return err
		}
		utxosAll = append(utxosAll, utxos...)
		if len(utxos) < int(maxEntries) {
			break
		}
		if runCount == 1 {
			i.services.Logger.Debugf("utxos %v", utxos)
			i.services.Logger.Debugf("endAddr %v", endAddr)
			i.services.Logger.Debugf("endUtxoID %v", endUtxoID)
		}

		//i.services.Logger.Debugf("run %v found %v utxos", runCount, len(utxosAll))

		runCount++
		nextStartAddr = endAddr
		nextStartUtxoID = endUtxoID
	}

	var txIDs []ids.ID
	for _, bytes := range utxosAll {
		utxo := &avax.UTXO{}
		_, err := txs.Codec.Unmarshal(bytes, utxo)
		if err != nil {
			return err
		}
		if utxo.TxID != ids.Empty {
			// Some utxos have empty TxID. Why?
			txIDs = append(txIDs, utxo.TxID)
		}
	}

	for _, txId := range txIDs {
		txBytes, err := client.GetTx(context.Background(), txId)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("get tx error %v", txId.String()))
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
