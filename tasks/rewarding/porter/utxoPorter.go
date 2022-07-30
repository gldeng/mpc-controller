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
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/secp256k1r"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	portIssuer "github.com/avalido/mpc-controller/utils/port/issuer"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

const (
	exportPrincipalTaskIDPrefix = "PRINCIPAL-"
	exportRewardTaskIDPrefix    = "REWARD-"
)

// Accept event: *events.ExportUTXORequestEvent

// Emit event: *events.UTXOExportedEvent

type UTXOPorter struct {
	CChainIssueClient chain.CChainIssuer
	Cache             cache.ICache
	Logger            logger.Logger
	MyPubKeyHashHex   string
	PChainIssueClient chain.PChainIssuer
	Publisher         dispatcher.Publisher
	SignDoner         core.SignDoner
	chain.NetworkContext

	evtObjChan             chan *dispatcher.EventObject
	UTXOsFetchedEventCache cache.SyncMapCache

	once                   sync.Once
	exportedRewardUTXOs    uint64
	exportedPrincipalUTXOs uint64
}

func (eh *UTXOPorter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.evtObjChan = make(chan *dispatcher.EventObject, 1024)
		go eh.exportUTXO(ctx)
	})

	select {
	case <-ctx.Done():
		return
	case eh.evtObjChan <- evtObj:
	}
}

func (eh *UTXOPorter) exportUTXO(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-eh.evtObjChan:
			evt, ok := evtObj.Event.(*events.ExportUTXORequestEvent)
			if !ok {
				break
			}

			partiKeys, err := eh.Cache.GetNormalizedParticipantKeys(evt.GenPubKeyHash, evt.ParticipantIndices)
			if err != nil { // todo: deal with err case
				eh.Logger.Error("UTXOPorter failed to export reward", []logger.Field{
					{"error", err},
					{"exportUTXORequestEvent", evt}}...)
				return
			}

			genPubKey := evt.GenPubKeyHash.Hex()
			val, ok := eh.UTXOsFetchedEventCache.Load(genPubKey)
			if !ok {
				eh.Logger.Warn("UTXOsFetchedEventCache not cached", []logger.Field{{"genPubKey", genPubKey}}...)
				return
			}
			utxoFectchedEvtObj := val.(*dispatcher.EventObject)
			utxoFetchedEvt := utxoFectchedEvtObj.Event.(*events.UTXOsFetchedEvent)

			var utxo *avax.UTXO
			txID := evt.TxID
			outputIndex := evt.OutputIndex
			for _, utxoFetched := range utxoFetchedEvt.NativeUTXOs {
				if utxoFetched.TxID == txID && utxoFetched.OutputIndex == outputIndex {
					utxo = utxoFetched
					break
				}
			}

			if utxo == nil {
				eh.Logger.Warn("UTXO not cached", []logger.Field{{"txID", txID}, {"outputIndex", outputIndex}}...)
				return
			}

			compressedGenPubKey, err := crypto.NormalizePubKey(utxoFetchedEvt.GenPubKeyHex)
			if err != nil {
				eh.Logger.Error("Failed to normalize generated public key", []logger.Field{
					{"error", err},
					{"genPubKey", utxoFetchedEvt.GenPubKeyHex}}...)
				return
			}

			taskID := exportPrincipalTaskIDPrefix
			if utxo.OutputIndex == 1 {
				taskID = exportRewardTaskIDPrefix
			}

			args := Args{
				Logger:     eh.Logger,
				NetworkID:  eh.NetworkID(),
				ExportFee:  eh.ExportFee(),
				PChainID:   ids.Empty, // todo: config it
				CChainID:   eh.CChainID(),
				PChainAddr: utxoFetchedEvt.PChainAddress,
				CChainArr:  evt.To,
				UTXO:       utxo,

				SignDoner: eh.SignDoner,
				SignReqArgs: &signer.SignRequestArgs{
					TaskID:                 taskID + evt.TxHash.Hex(),
					CompressedPartiPubKeys: partiKeys,
					CompressedGenPubKeyHex: *compressedGenPubKey,
				},

				CChainIssueClient: eh.CChainIssueClient,
				PChainIssueClient: eh.PChainIssueClient,
			}

			ids, err := doExportUTXO(ctx, &args)
			if err != nil {
				switch errors.Cause(err).(type) { // todo: exploring more concrete error types
				case *chain.ErrTypSharedMemoryNotFound:
					eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
						{"txID", evt.TxID},
						{"outputIndex", evt.OutputIndex}}...)
				case *chain.ErrTypConflictAtomicInputs:
					eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
						{"txID", evt.TxID},
						{"outputIndex", evt.OutputIndex}}...)
				case *chain.ErrTypImportUTXOsNotFound:
					eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
						{"txID", evt.TxID},
						{"outputIndex", evt.OutputIndex}}...)
				case *chain.ErrTypConsumedUTXONotFound:
					eh.Logger.DebugOnError(err, "UTXO UNEXPORTED", []logger.Field{
						{"txID", evt.TxID},
						{"outputIndex", evt.OutputIndex}}...)
				default:
					eh.Logger.ErrorOnError(err, "Failed to export UTXO", []logger.Field{
						{"txID", evt.TxID},
						{"outputIndex", evt.OutputIndex}}...)
				}
				return
			}

			mpcUTXO := myAvax.MpcUTXOFromUTXO(utxo)
			newEvt := &events.UTXOExportedEvent{
				NativeUTXO:   utxo,
				MpcUTXO:      mpcUTXO,
				ExportedTxID: ids[0],
				ImportedTxID: ids[1],
			}
			eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "ExportUTXORequestWatcher", newEvt, evtObj.Context))

			switch utxo.OutputIndex {
			case 0:
				atomic.AddUint64(&eh.exportedPrincipalUTXOs, 1)
				eh.Logger.Info("Principal UTXO EXPORTED", []logger.Field{{"UTXOExportedEvent", newEvt}}...)
			case 1:
				atomic.AddUint64(&eh.exportedRewardUTXOs, 1)
				eh.Logger.Info("Reward UTXO EXPORTED", []logger.Field{{"UTXOExportedEvent", newEvt}}...)
			}
			totalPrincipals := atomic.LoadUint64(&eh.exportedPrincipalUTXOs)
			totalRewards := atomic.LoadUint64(&eh.exportedRewardUTXOs)
			eh.Logger.Info("Exported UTXO stats", []logger.Field{{"exportedPrincipalUTXOs", totalPrincipals}, {"exportedRewardUTXOs", totalRewards}}...)

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
