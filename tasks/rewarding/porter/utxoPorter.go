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
)

// Accept event: *events.UTXOsFetchedEvent
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

	once sync.Once

	// todo: consider include this field in Cache or dispatcher
	utxoFetchedEvtObjMap map[string]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *UTXOPorter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.UTXOsFetchedEvent:
		eh.once.Do(func() {
			eh.utxoFetchedEvtObjMap = make(map[string]*dispatcher.EventObject)
		})
		eh.utxoFetchedEvtObjMap[evt.GenPubKeyHashHex] = evtObj
	case *events.ExportUTXORequestEvent:
		ok := eh.Cache.IsParticipant(eh.MyPubKeyHashHex, evt.GenPubKeyHash.Hex(), evt.ParticipantIndices)
		if !ok {
			eh.Logger.Debug("Not participated ExportUTXORequest", []logger.Field{{"exportUTXORequest", evt}}...)
			break
		}
		eh.exportUTXO(ctx, evtObj)
	}
}

func (eh *UTXOPorter) exportUTXO(ctx context.Context, evtObj *dispatcher.EventObject) {
	exportUTXOReqEvt := evtObj.Event.(*events.ExportUTXORequestEvent)

	partiKeys, err := eh.Cache.GetNormalizedParticipantKeys(exportUTXOReqEvt.GenPubKeyHash, exportUTXOReqEvt.ParticipantIndices)
	if err != nil { // todo: deal with err case
		eh.Logger.Error("UTXOPorter failed to export reward", []logger.Field{
			{"error", err},
			{"exportUTXORequestEvent", exportUTXOReqEvt}}...)
		return
	}

	utxoFectchedEvtObj := eh.utxoFetchedEvtObjMap[exportUTXOReqEvt.GenPubKeyHash.Hex()]
	utxoFetchedEvt := utxoFectchedEvtObj.Event.(*events.UTXOsFetchedEvent)

	var utxo *avax.UTXO
	for _, utxo = range utxoFetchedEvt.NativeUTXOs {
		if utxo.TxID == exportUTXOReqEvt.TxID && utxo.OutputIndex == exportUTXOReqEvt.OutputIndex {
			break
		}
	}

	if utxo == nil {
		eh.Logger.Warn("Found no UTXO to export", []logger.Field{{"exportUTXORequestEvent", exportUTXOReqEvt}}...)
		return
	}

	compressedGenPubKey, err := crypto.NormalizePubKey(utxoFetchedEvt.GenPubKeyHex)
	if err != nil {
		eh.Logger.Error("Failed to normalize generated public key", []logger.Field{
			{"error", err},
			{"genPubKey", utxoFetchedEvt.GenPubKeyHex}}...)
		return
	}

	args := Args{
		NetworkID: eh.NetworkID(),
		ExportFee: eh.ExportFee(),
		//PChainID // todo:
		CChainID:   eh.CChainID(),
		PChainAddr: utxoFetchedEvt.PChainAddress,
		CChainArr:  exportUTXOReqEvt.To,
		UTXO:       utxo,

		SignDoner: eh.SignDoner,
		SignReqArgs: &signer.SignRequestArgs{
			TaskID:                 exportUTXOReqEvt.TxHash.Hex(),
			CompressedPartiPubKeys: partiKeys,
			CompressedGenPubKeyHex: *compressedGenPubKey,
		},

		CChainIssueClient: eh.CChainIssueClient,
		PChainIssueClient: eh.PChainIssueClient,
	}

	ids, err := doExportUTXO(ctx, &args)
	if err != nil {
		eh.Logger.Error("UTXOPorter failed to export UTXO", []logger.Field{
			{"error", err},
			{"exportUTXORequestEvent", exportUTXOReqEvt}}...)
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
}

type Args struct {
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
	myExportTxArgs := &pchain.Args{
		NetworkID:          args.NetworkID,
		BlockchainID:       args.PChainID,
		DestinationChainID: args.CChainID,
		Amount:             amountToExport - args.ExportFee,
		To:                 args.PChainAddr,
		UTXOs:              []*avax.UTXO{args.UTXO},
	}

	myImportTxArgs := &cchain.Args{
		NetworkID:     args.NetworkID,
		BlockchainID:  args.CChainID,
		SourceChainID: args.PChainID,
		To:            args.CChainArr,
		//BaseFee:      *big.Int todo: tobe consider
	}

	myTxs := &Txs{
		UnsignedExportTxArgs: myExportTxArgs,
		UnsignedImportTx:     myImportTxArgs,
	}

	mySigner := &signer.Signer{args.SignDoner, *args.SignReqArgs}
	myVerifier := &secp256k1r.Verifier{PChainAddress: args.PChainAddr}
	myIssuer := &portIssuer.Issuer{args.CChainIssueClient, args.PChainIssueClient, portIssuer.P2C}
	myPorter := porter.Porter{myTxs, mySigner, myIssuer, myVerifier}

	txIds, err := myPorter.SignAndIssueTxs(ctx)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	return txIds, nil
}
