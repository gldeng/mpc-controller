package tracker

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"time"
)

//type Cache interface {
//	cache.MyIndexGetter
//	cache.GeneratedPubKeyInfoGetter
//	cache.ParticipantKeysGetter
//}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding reward

type RewardUTXOGetter interface {
	GetRewardUTXOs(context.Context, *api.GetTxArgs, ...rpc.Option) ([][]byte, error)
}

// Accept event: *events.StakingTaskDoneEvent

// Emit event:

type RewardUTXOTracker struct {
	Logger logger.Logger
	//chain.NetworkContext
	//
	//MyPubKeyHashHex string
	//
	//Cache Cache
	//
	//SignDoner core.SignDoner
	//Publisher dispatcher.Publisher
	//
	//CChainIssueClient chain.Issuer
	//PChainIssueClient chain.Issuer
	//
	//Noncer chain.Noncer
	//
	//genPubKeyInfo *events.GeneratedPubKeyInfo
	//myIndex       *big.Int
	RewardUTXOGetter RewardUTXOGetter
}

func (eh *RewardUTXOTracker) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.StakingTaskDoneEvent:
		eh.do(evt, evtObj)
	}
}

func (eh *RewardUTXOTracker) do(evt *events.StakingTaskDoneEvent, evtObj *dispatcher.EventObject) {
	utxoArr, err := eh.requestRewardUTXOs(evtObj.Context, evt.AddDelegatorTxID)
	if err != nil {
		eh.Logger.Error("Failed to request reward UTXO for addDelegatorTx %q", []logger.Field{
			{"error", err},
			{"addDelegatorTx", evt.AddDelegatorTxID}}...)
	}

	time.Sleep(time.Second * 12)

	fmt.Println("length of utxoArr: ", len(utxoArr))
	for _, utxo := range utxoArr {
		output := utxo.Out.(*secp256k1fx.TransferOutput)
		spew.Dump(utxo)
		spew.Dump(output)
	}
}

func (eh *RewardUTXOTracker) requestRewardUTXOs(ctx context.Context, txID ids.ID) ([]*avax.UTXO, error) {
	utxosBytesArr, err := eh.RewardUTXOGetter.GetRewardUTXOs(ctx, &api.GetTxArgs{TxID: txID})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var utxos = make([]*avax.UTXO, len(utxosBytesArr))

	for _, utxoBytes := range utxosBytesArr {
		var utxo avax.UTXO
		_, err := platformvm.Codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		utxos = append(utxos, &utxo)
	}

	return utxos, nil
}
