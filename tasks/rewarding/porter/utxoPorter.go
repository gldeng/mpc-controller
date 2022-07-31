package porter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/signer"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/secp256k1r"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	portIssuer "github.com/avalido/mpc-controller/utils/port/issuer"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/avalido/mpc-controller/utils/work"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	exportPrincipalTaskIDPrefix = "PRINCIPAL-"
	exportRewardTaskIDPrefix    = "REWARD-"
)

// Subscribe event: *events.ReportedGenPubKeyEvent
// Subscribe event: *events.UTXOReportedEvent
// Subscribe event: *events.ExportUTXORequestEvent

// Publish event: *events.UTXOExportedEvent

type UTXOPorter struct {
	CChainIssueClient chain.CChainIssuer
	Cache             cache.ICache
	Logger            logger.Logger
	MyPubKeyHashHex   string
	PChainIssueClient chain.PChainIssuer
	Publisher         dispatcher.Publisher
	SignDoner         core.SignDoner
	chain.NetworkContext

	ws *work.Workshop

	reportedGenPubKeyEventCache map[string]*events.ReportedGenPubKeyEvent
	utxoReportedEventCache      map[string]*events.UTXOReportedEvent // todo: obsolete, use UTXOReportedEventCache instead?
	UTXOReportedEventCache      *sync.Map

	ExportUTXORequestEventChan chan *events.ExportUTXORequestEvent

	once                   sync.Once
	exportedRewardUTXOs    uint64
	exportedPrincipalUTXOs uint64
}

func (eh *UTXOPorter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.reportedGenPubKeyEventCache = make(map[string]*events.ReportedGenPubKeyEvent)
		eh.utxoReportedEventCache = make(map[string]*events.UTXOReportedEvent)

		eh.ExportUTXORequestEventChan = make(chan *events.ExportUTXORequestEvent, 1024)
		eh.ws = work.NewWorkshop(eh.Logger, "signRewardTx", time.Minute*10, 10)

		go eh.exportUTXO(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.ReportedGenPubKeyEvent:
		eh.reportedGenPubKeyEventCache[evt.GenPubKeyHashHex] = evt
	case *events.UTXOReportedEvent:
		utxoID := evt.NativeUTXO.TxID.String() + strconv.Itoa(int(evt.NativeUTXO.OutputIndex))
		eh.utxoReportedEventCache[utxoID] = evt
	case *events.ExportUTXORequestEvent:
		select {
		case <-ctx.Done():
			return
		case eh.ExportUTXORequestEventChan <- evt:
		}
	}
}

func (eh *UTXOPorter) exportUTXO(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-eh.ExportUTXORequestEventChan:
			partiKeys, err := eh.Cache.GetNormalizedParticipantKeys(evt.GenPubKeyHash, evt.ParticipantIndices)
			if err != nil { // todo: deal with err case
				eh.Logger.Error("UTXOPorter failed to export reward", []logger.Field{
					{"error", err},
					{"exportUTXORequestEvent", evt}}...)
				break
			}

			utxoID := evt.TxID.String() + strconv.Itoa(int(evt.OutputIndex))
			val, ok := eh.UTXOReportedEventCache.Load(utxoID)
			if !ok {
				eh.Logger.Warn("No local reported UTXO found", []logger.Field{
					{"txID", evt.TxID}, {"outputIndex", evt.OutputIndex}}...)
				break
			}
			utxoRepEvt := val.(*events.UTXOReportedEvent)

			//utxoRepEvt, ok := eh.utxoReportedEventCache[utxoID]
			//if !ok || utxoRepEvt == nil {
			//	eh.Logger.Warn("No local reported UTXO found", []logger.Field{
			//		{"txID", evt.TxID}, {"outputIndex", evt.OutputIndex}}...)
			//	break
			//}

			genPubKeyHex := bytes.BytesToHex(utxoRepEvt.GenPubKeyBytes)
			compressedGenPubKey, err := crypto.NormalizePubKey(genPubKeyHex)
			if err != nil {
				eh.Logger.Error("Failed to normalize generated public key", []logger.Field{
					{"error", err},
					{"genPubKey", genPubKeyHex}}...)
				break
			}

			genPubKeyEvt := eh.reportedGenPubKeyEventCache[evt.GenPubKeyHash.String()]

			taskID := exportPrincipalTaskIDPrefix
			if utxoRepEvt.NativeUTXO.OutputIndex == 1 {
				taskID = exportRewardTaskIDPrefix
			}

			args := &Args{
				Logger:     eh.Logger,
				NetworkID:  eh.NetworkID(),
				ExportFee:  eh.ExportFee(),
				PChainID:   ids.Empty, // todo: config it
				CChainID:   eh.CChainID(),
				PChainAddr: genPubKeyEvt.PChainAddress,
				CChainArr:  evt.To,
				UTXO:       utxoRepEvt.NativeUTXO,

				SignDoner: eh.SignDoner,
				SignReqArgs: &signer.SignRequestArgs{
					TaskID:                 taskID + evt.TxHash.Hex(),
					CompressedPartiPubKeys: partiKeys,
					CompressedGenPubKeyHex: *compressedGenPubKey,
				},

				CChainIssueClient: eh.CChainIssueClient,
				PChainIssueClient: eh.PChainIssueClient,
			}

			eh.ws.AddTask(ctx, &work.Task{
				Args: args,
				Ctx:  ctx,
				WorkFns: []work.WorkFn{func(ctx context.Context, args interface{}) {
					argsVal := args.(*Args)
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

					newEvt := &events.UTXOExportedEvent{
						NativeUTXO:   utxoRepEvt.NativeUTXO,
						MpcUTXO:      utxoRepEvt.MpcUTXO,
						ExportedTxID: ids[0],
						ImportedTxID: ids[1],
					}
					eh.Publisher.Publish(ctx, dispatcher.NewEvtObj(newEvt, nil))

					switch utxoRepEvt.NativeUTXO.OutputIndex {
					case 0:
						atomic.AddUint64(&eh.exportedPrincipalUTXOs, 1)
						eh.Logger.Info("Principal UTXO EXPORTED", []logger.Field{{"UTXOExportedEvent", newEvt}}...)
					case 1:
						atomic.AddUint64(&eh.exportedRewardUTXOs, 1)
						eh.Logger.Info("Reward UTXO EXPORTED", []logger.Field{{"UTXOExportedEvent", newEvt}}...)
					}
					totalPrincipals := atomic.LoadUint64(&eh.exportedPrincipalUTXOs)
					totalRewards := atomic.LoadUint64(&eh.exportedRewardUTXOs)
					eh.Logger.Info("Exported UTXO stats", []logger.Field{{"exportedPrincipalUTXOs", totalPrincipals},
						{"exportedRewardUTXOs", totalRewards}}...)
				}},
			})
		}
	}
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
