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
	"github.com/avalido/mpc-controller/utils/bytes"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/pkg/errors"
	"github.com/status-im/keycard-go/hexutils"
	"sort"
	"sync"
	"time"
)

type Queue interface {
	Enqueue(value interface{}) error
}

type Indexer struct {
	services         *core.ServicePack
	eventQueue       Queue
	closeOnce        sync.Once
	onCloseCtx       context.Context
	onCloseCtxCancel func()
	bucketsSent      map[int64]struct{}
}

func NewIndexer(services *core.ServicePack, eventQueue Queue) *Indexer {
	onCloseCtx, cancel := context.WithCancel(context.Background())
	return &Indexer{
		services:         services,
		eventQueue:       eventQueue,
		closeOnce:        sync.Once{},
		onCloseCtx:       onCloseCtx,
		onCloseCtxCancel: cancel,
		bucketsSent:      map[int64]struct{}{},
	}
}

func (i *Indexer) Start() error {
	client := i.services.Config.CreatePClient()

	go func() {

		timer := time.NewTimer(core.DefaultParameters.IndexerStartDelay)
		defer timer.Stop()

		for {
			select {
			case <-i.onCloseCtx.Done():
				return
			case <-timer.C:
				nextRun := time.Now().Add(core.DefaultParameters.IndexerLoopDuration)
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
				i.services.TxIndex.PurgeOlderThan(time.Now().Add(-core.DefaultParameters.TxIndexPurgueAge))
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
		err = i.services.TxIndex.SetTxByType(reqHash, core.TxTypeAddDelegator, txId)
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

	utxosByTxId, err := GroupUtxosByTxId(utxosAll)
	if err != nil {
		return err
	}

	addDelegatorTxs, err := i.getRequestHashAndUpdateIndex(client, utxosByTxId)
	if err != nil {
		return err
	}

	bucket := getCurrentBucket()
	if _, ok := i.bucketsSent[bucket.EndTimestamp]; ok {
		return nil
	}

	inBuckeTxs := findTxsInTargetBucket(addDelegatorTxs, bucket)
	byPubKey, err := groupByPubKey(inBuckeTxs)
	if err != nil {
		return err
	}

	principals, rewards, err := splitUtxosByType(utxosByTxId, *byPubKey)
	if err != nil {
		return err
	}

	var events []types.UtxoBucket
	for pkStr, utxos := range principals {
		events = append(events, types.UtxoBucket{
			StartTimestamp: uint64(bucket.StartTimestamp),
			EndTimestamp:   uint64(bucket.EndTimestamp),
			PublicKey:      hexutils.HexToBytes(pkStr),
			Utxos:          utxos,
			UtxoType:       types.Principal,
		})
	}
	for pkStr, utxos := range rewards {
		events = append(events, types.UtxoBucket{
			StartTimestamp: uint64(bucket.StartTimestamp),
			EndTimestamp:   uint64(bucket.EndTimestamp),
			PublicKey:      hexutils.HexToBytes(pkStr),
			Utxos:          utxos,
			UtxoType:       types.Reward,
		})
	}
	for _, event := range events {
		i.eventQueue.Enqueue(event)
	}
	i.bucketsSent[bucket.EndTimestamp] = struct{}{}

	return nil
}

func (i *Indexer) getRequestHashAndUpdateIndex(client platformvm.Client, container UtxosByTxId) ([]*txs.Tx, error) {
	var allAddDelegatorTxs []*txs.Tx
	for txId, _ := range container {
		txBytes, err := client.GetTx(context.Background(), txId)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("get tx error %v", txId.String()))
		}
		tx, err := txs.Parse(txs.Codec, txBytes)
		if err != nil {
			return nil, err
		}
		tx.Sign(txs.Codec, nil)
		if uTx, ok := tx.Unsigned.(*txs.ImportTx); ok {
			reqHash := types.RequestHash{}
			copy(reqHash[:], uTx.Memo)
			err = i.services.TxIndex.SetTxByType(reqHash, core.TxTypeImportP, txId)
			if err != nil {
				return nil, err
			}
		} else if uTx, ok := tx.Unsigned.(*txs.AddDelegatorTx); ok {
			reqHash := types.RequestHash{}
			copy(reqHash[:], uTx.Memo)
			err = i.services.TxIndex.SetTxByType(reqHash, core.TxTypeAddDelegator, txId)
			if err != nil {
				return nil, err
			}
			allAddDelegatorTxs = append(allAddDelegatorTxs, tx)
		}
	}
	return allAddDelegatorTxs, nil
}

func findTxsInTargetBucket(transactions []*txs.Tx, bucket Bucket) []*txs.Tx {
	var inBucketTxs []*txs.Tx
	for _, tx := range transactions {
		uTx, _ := tx.Unsigned.(*txs.AddDelegatorTx)
		if bucket.isInBucket(uTx.EndTime()) {
			inBucketTxs = append(inBucketTxs, tx)
		}
	}
	return inBucketTxs
}

func getCurrentBucket() Bucket {
	now := time.Now().Unix()
	endTimestamp := (now / core.DefaultParameters.UtxoBucketSeconds) * core.DefaultParameters.UtxoBucketSeconds
	startTimestamp := endTimestamp - core.DefaultParameters.UtxoBucketSeconds
	return Bucket{
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}
}

func groupByPubKey(transactions []*txs.Tx) (*ByPubKey, error) {
	byPubKey := ByPubKey{}
	for _, tx := range transactions {
		uTx, _ := tx.Unsigned.(*txs.AddDelegatorTx)
		pk, err := myCrypto.RecoverPChainTxPublicKey(tx)
		if err != nil {
			return nil, err
		}
		pkStr := bytes.BytesToHex(pk)
		byPubKey[pkStr] = append(byPubKey[pkStr], AddDelegatorTxRecord{
			TxID:    tx.ID(),
			EndTime: uTx.EndTime().Unix(),
		})
	}

	for _, records := range byPubKey {
		sort.Sort(records)
	}
	return &byPubKey, nil
}

func splitUtxosByType(utxosByTxId UtxosByTxId, byPubKey ByPubKey) (principals UtxosByPubKey, rewards UtxosByPubKey, err error) {
	principals = map[string][]*avax.UTXO{}
	rewards = map[string][]*avax.UTXO{}
	for pkStr, records := range byPubKey {
		for _, rec := range records {
			utxos, ok := utxosByTxId[rec.TxID]
			if !ok {
				return nil, nil, errors.New(fmt.Sprintf("failed to find utxo for txID %v", rec.TxID.String()))
			}
			sort.Sort(SortUtxos(utxos))
			principals[pkStr] = append(principals[pkStr], utxos[0])
			rewards[pkStr] = append(rewards[pkStr], utxos[1:]...)
		}
	}
	return principals, rewards, nil
}

func GroupUtxosByTxId(utxosAll [][]byte) (UtxosByTxId, error) {
	utxosByTxId := UtxosByTxId{}

	for _, bytes := range utxosAll {
		utxo := &avax.UTXO{}
		_, err := txs.Codec.Unmarshal(bytes, utxo)
		if err != nil {
			return nil, err
		}
		if utxo.TxID != ids.Empty {
			if _, ok := utxosByTxId[utxo.TxID]; !ok {
				utxosByTxId[utxo.TxID] = nil
			}
			utxosByTxId[utxo.TxID] = append(utxosByTxId[utxo.TxID], utxo)
		}
	}
	return utxosByTxId, nil
}

type UtxosByTxId map[ids.ID][]*avax.UTXO
type SortUtxos []*avax.UTXO

func (a SortUtxos) Len() int      { return len(a) }
func (a SortUtxos) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortUtxos) Less(i, j int) bool {
	return a[i].OutputIndex < a[j].OutputIndex
}

type Bucket struct {
	StartTimestamp int64
	EndTimestamp   int64
}

func (b *Bucket) isInBucket(ti time.Time) bool {
	ts := ti.Unix()
	return ts >= b.StartTimestamp && ts < b.EndTimestamp
}

type AddDelegatorTxRecord struct {
	TxID    ids.ID
	EndTime int64
}

type SortAddDelegatorTxRecord []AddDelegatorTxRecord

func (a SortAddDelegatorTxRecord) Len() int      { return len(a) }
func (a SortAddDelegatorTxRecord) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortAddDelegatorTxRecord) Less(i, j int) bool {
	if a[i].EndTime < a[j].EndTime {
		return true
	}
	return a[i].TxID.Hex() < a[j].TxID.Hex()
}

type ByPubKey = map[string]SortAddDelegatorTxRecord
type UtxosByPubKey = map[string][]*avax.UTXO
