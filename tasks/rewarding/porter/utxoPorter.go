package porter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/signer"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/crypto/secp256k1r"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	portIssuer "github.com/avalido/mpc-controller/utils/port/issuer"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/avalido/mpc-controller/utils/work"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	taskIDPrefixPrincipalExport = "PRINCIPAL-"
	taskIDPrefixRewardExport    = "REWARD-"
)

const (
	utxoOutputIndexPrincipal utxoOutputIndex = iota
	utxoOutputIndexReward
)

type utxoOutputIndex int

// Subscribe event: *events.RequestStarted

type UTXOPorter struct {
	BoundCaller            caller.Caller
	DB                     storage.DB
	IssuerCChain           chain.CChainIssuer
	IssuerPChain           chain.PChainIssuer
	Logger                 logger.Logger
	NetWorkCtx             chain.NetworkContext
	SignerMPC              core.SignDoner
	UTXOExportedEventCache *ristretto.Cache
	UTXOsFetchedEventCache *ristretto.Cache

	ws                       *work.Workshop
	requestStartedChan       chan *events.RequestStarted
	exportUTXOTaskAddedCache *ristretto.Cache
	once                     sync.Once
	exportedRewardUTXOs      uint64
	exportedPrincipalUTXOs   uint64
}

func (eh *UTXOPorter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.requestStartedChan = make(chan *events.RequestStarted, 1024)
		eh.ws = work.NewWorkshop(eh.Logger, "signRewardTx", time.Minute*10, 10)

		exportUTXOTaskAddedCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		eh.exportUTXOTaskAddedCache = exportUTXOTaskAddedCache

		go eh.exportUTXO(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.RequestStarted:
		reqHash := (storage.RequestHash)(evt.RequestHash)
		if !reqHash.IsTaskType(storage.TaskTypReturn) {
			break
		}
		select {
		case <-ctx.Done():
			return
		case eh.requestStartedChan <- evt:
		}
	}
}

func (eh *UTXOPorter) exportUTXO(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-eh.requestStartedChan:
			reqHash := (storage.RequestHash)(evt.RequestHash)
			utxoExportReq := storage.ExportUTXORequest{}
			joinReq := storage.JoinRequest{
				ReqHash: reqHash,
				Args:    &utxoExportReq,
			}

			if err := eh.DB.LoadModel(ctx, &joinReq); err != nil {
				eh.Logger.DebugOnError(err, "No JoinRequest load for UTXO export", []logger.Field{{"reqHash", joinReq.ReqHash}}...)
				break
			}
			if !joinReq.PartiId.Joined(evt.ParticipantIndices) {
				eh.Logger.Debug("Not joined UTXO export request", []logger.Field{{"reqHash", evt.RequestHash}}...)
				break
			}

			group := storage.Group{
				ID: utxoExportReq.GroupId,
			}
			if err := eh.DB.LoadModel(ctx, &group); err != nil {
				eh.Logger.ErrorOnError(err, "Failed to load group", []logger.Field{{"key", group.ID}}...)
				break
			}

			cmpPartiPubKeys, err := group.Group.CompressPubKeyHexs()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to compress participant public keys")
				break
			}

			cmpGenPubKeyHex, err := utxoExportReq.GenPubKey.CompressPubKeyHex()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to compress generated public key")
				break
			}

			pChainAddr, err := utxoExportReq.GenPubKey.PChainAddress()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to get P-Chain address")
				break
			}

			treasureAddr, err := eh.treasuryAddress(ctx, utxoOutputIndex(utxoExportReq.OutputIndex))
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to get treasure address")
				break
			}

			utxoID := utxoExportReq.TxID.String() + strconv.Itoa(int(utxoExportReq.OutputIndex))
			_, ok := eh.exportUTXOTaskAddedCache.Get(utxoID)
			if ok {
				break
			}
			val, ok := eh.UTXOsFetchedEventCache.Get(utxoID)
			if !ok {
				eh.Logger.Warn("No local reported UTXO found", []logger.Field{
					{"txID", utxoExportReq.TxID}, {"outputIndex", utxoExportReq.OutputIndex}}...)
				break
			}
			utxo := val.(*avax.UTXO)

			taskID := taskIDPrefixPrincipalExport
			if utxo.OutputIndex == 1 {
				taskID = taskIDPrefixRewardExport
			}

			args := &Args{
				Logger:     eh.Logger,
				NetworkID:  eh.NetWorkCtx.NetworkID(),
				ExportFee:  eh.NetWorkCtx.ExportFee(),
				PChainID:   ids.Empty,
				CChainID:   eh.NetWorkCtx.CChainID(),
				PChainAddr: pChainAddr,
				CChainArr:  treasureAddr,
				UTXO:       utxo,

				SignDoner: eh.SignerMPC,
				SignReqArgs: &signer.SignRequestArgs{
					TaskID:                 taskID + evt.Raw.TxHash.Hex(),
					CompressedPartiPubKeys: cmpPartiPubKeys,
					CompressedGenPubKeyHex: cmpGenPubKeyHex,
				},

				CChainIssueClient: eh.IssuerCChain,
				PChainIssueClient: eh.IssuerPChain,
			}

			eh.exportUTXOTaskAddedCache.SetWithTTL(utxoID, " ", 1, time.Hour)
			eh.exportUTXOTaskAddedCache.Wait()

			err = eh.ws.AddTask(ctx, &work.Task{
				Args: args,
				Ctx:  ctx,
				WorkFns: []work.WorkFn{func(ctx context.Context, args interface{}) {
					argsVal := args.(*Args)
					utxo := argsVal.UTXO
					utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
					_, ok := eh.UTXOExportedEventCache.Get(utxoID)
					if ok {
						return
					}

					eh.Logger.Debug("Starting exporting UTXO task...", []logger.Field{
						{"taskID", argsVal.SignReqArgs.TaskID},
						{"utxoID", argsVal.UTXO.UTXOID}}...)
					ids, err := doExportUTXO(ctx, argsVal)
					if err != nil {
						switch errors.Cause(err).(type) { // todo: exploring more concrete error types
						case *chain.ErrTypSharedMemoryNotFound:
							eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
								{"taskID", argsVal.SignReqArgs.TaskID},
								{"utxoID", argsVal.UTXO.UTXOID}}...)
						case *chain.ErrTypConflictAtomicInputs:
							eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
								{"taskID", argsVal.SignReqArgs.TaskID},
								{"utxoID", argsVal.UTXO.UTXOID}}...)
						case *chain.ErrTypImportUTXOsNotFound:
							eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
								{"taskID", argsVal.SignReqArgs.TaskID},
								{"utxoID", argsVal.UTXO.UTXOID}}...)
						case *chain.ErrTypConsumedUTXONotFound:
							eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
								{"taskID", argsVal.SignReqArgs.TaskID},
								{"utxoID", argsVal.UTXO.UTXOID}}...)
						default:
							eh.Logger.ErrorOnError(err, "Failed to export UTXO", []logger.Field{
								{"taskID", argsVal.SignReqArgs.TaskID},
								{"utxoID", argsVal.UTXO.UTXOID}}...)
						}
						return
					}

					newEvt := &events.UTXOExported{
						NativeUTXO:   utxo,
						MpcUTXO:      myAvax.MpcUTXOFromUTXO(utxo),
						ExportedTxID: ids[0],
						ImportedTxID: ids[1],
					}

					eh.UTXOExportedEventCache.SetWithTTL(utxoID, " ", 1, time.Hour)
					eh.UTXOExportedEventCache.Wait()

					prom.UTXOExported.Inc()
					switch utxo.OutputIndex {
					case uint32(events.OutputIndexPrincipal):
						atomic.AddUint64(&eh.exportedPrincipalUTXOs, 1)
						prom.PrincipalUTXOExported.Inc()
						eh.Logger.Info("Principal UTXO EXPORTED", []logger.Field{{"UTXOExported", newEvt}}...)
					case uint32(events.OutputIndexReward):
						atomic.AddUint64(&eh.exportedRewardUTXOs, 1)
						prom.RewardUTXOExported.Inc()
						eh.Logger.Info("Reward UTXO EXPORTED", []logger.Field{{"UTXOExported", newEvt}}...)
					}
					totalPrincipals := atomic.LoadUint64(&eh.exportedPrincipalUTXOs)
					totalRewards := atomic.LoadUint64(&eh.exportedRewardUTXOs)
					eh.Logger.Info("Exported UTXO stats", []logger.Field{{"exportedPrincipalUTXOs", totalPrincipals},
						{"exportedRewardUTXOs", totalRewards}}...)
				}},
			})
			eh.Logger.ErrorOnError(err, "Failed to add UTXO export task", []logger.Field{
				{"reqHash", reqHash.String()}, {"taskId", taskID}, {"txID", utxo.TxID}, {"outputIndex", utxo.OutputIndex}}...)
			eh.Logger.InfoNilError(err, "UTXO export task added", []logger.Field{
				{"reqHash", reqHash.String()}, {"taskId", taskID}, {"txID", utxo.TxID}, {"outputIndex", utxo.OutputIndex}}...)
		}
	}
}

func (eh *UTXOPorter) treasuryAddress(ctx context.Context, outputIndex utxoOutputIndex) (addr common.Address, err error) {
	switch {
	case outputIndex == utxoOutputIndexPrincipal:
		if addr, err = eh.BoundCaller.PrincipalTreasuryAddress(ctx, nil); err != nil {
			return *new(common.Address), errors.WithStack(err)
		}

	case outputIndex == utxoOutputIndexReward:
		if addr, err = eh.BoundCaller.RewardTreasuryAddress(ctx, nil); err != nil {
			return *new(common.Address), errors.WithStack(err)
		}
	}
	return
}

type Args struct {
	Logger     logger.Logger
	NetworkID  uint32
	ExportFee  uint64
	PChainID   ids.ID
	CChainID   ids.ID
	PChainAddr ids.ShortID
	CChainArr  common.Address
	UTXO       *avax.UTXO

	SignDoner   core.SignDoner
	SignReqArgs *signer.SignRequestArgs

	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer
}

func doExportUTXO(ctx context.Context, args *Args) ([2]ids.ID, error) {
	amountToExport := args.UTXO.Out.(*secp256k1fx.TransferOutput).Amount()
	if amountToExport < args.ExportFee {
		return [2]ids.ID{}, errors.Errorf("amoutToExport(%v) is less than exportFee(%v)", amountToExport, args.ExportFee)
	}
	outAmount := amountToExport - args.ExportFee // todo: consider batch export to reduce fee
	myExportTxArgs := &pchain.ExportTxArgs{
		NetworkID:          args.NetworkID,
		BlockchainID:       args.PChainID,
		DestinationChainID: args.CChainID,
		OutAmount:          outAmount,
		To:                 args.PChainAddr,
		UTXOs:              []*avax.UTXO{args.UTXO},
	}
	myImportTxArgs := &cchain.ImportTxArgs{
		NetworkID:     args.NetworkID,
		BlockchainID:  args.CChainID,
		OutAmount:     outAmount,
		SourceChainID: args.PChainID,
		To:            args.CChainArr,
	}

	myTxs := &Txs{
		ExportTxArgs: myExportTxArgs,
		ImportTxArgs: myImportTxArgs,
	}

	mySigner := &signer.Signer{args.SignDoner, *args.SignReqArgs}
	myVerifier := &secp256k1r.Verifier{PChainAddress: args.PChainAddr}
	myIssuer := &portIssuer.Issuer{args.CChainIssueClient, args.PChainIssueClient, portIssuer.P2C}
	myPorter := porter.Porter{args.Logger, myTxs, mySigner, myIssuer, myVerifier}

	txIds, err := myPorter.SignAndIssueTxs(ctx)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}
	return txIds, nil
}
