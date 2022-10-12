package stake

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"math/big"
)

type TaskCreator struct {
	Ctx    context.Context
	Logger logger.Logger

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	NonceGiver noncer.Noncer
	Network    chain.NetworkContext

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.RequestStarted]
}

func (c *TaskCreator) Init() {
	reqStartedEvtHandler := func(evt *events.RequestStarted) {
		task, err := c.createTask(evt)
		if err != nil {
			c.Logger.ErrorOnError(err, "Failed to create task")
			return
		}
		c.Pool.Submit(task.Do)
	}

	reqStartedEvtFilter := func(evt *events.RequestStarted) bool {
		return evt.TaskType == storage.TaskTypStake
	}

	c.Dispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
}

func (c *TaskCreator) createTask(joined *events.RequestStarted) (*Task, error) {
	stakeReq := joined.JoinedReq.Args.(*storage.StakeRequest)
	txs, err := c.createTxs(stakeReq, *joined.ReqHash)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signReqs, err := c.createSignReqs(joined, txs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t := Task{
		Ctx:    c.Ctx,
		Logger: c.Logger,

		MpcClient: c.MpcClient,
		TxIssuer:  c.TxIssuer,

		Pool:       c.Pool,
		Dispatcher: kbcevents.NewDispatcher[*events.StakeAtomicTaskDone](),

		Txs:             txs,
		ExportTxSignReq: signReqs[0],
		ImportTxSignReq: signReqs[1],
	}

	return &t, nil
}

func (c *TaskCreator) createSignReqs(joined *events.RequestStarted, txs *Txs) ([]*core.SignRequest, error) {
	exportTxHash, err := txs.ExportTxHash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	exportTxSignReq := core.SignRequest{
		ReqID:                  string(events.SignIDPrefixStakeExport) + fmt.Sprintf("%v", txs.ReqNo) + "-" + txs.ReqHash,
		Kind:                   events.SignKindStakeExport,
		CompressedGenPubKeyHex: joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	importTxHash, err := txs.ImportTxHash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	importTxSignReq := core.SignRequest{
		ReqID:                  string(events.SignIDPrefixStakeImport) + fmt.Sprintf("%v", txs.ReqNo) + "-" + txs.ReqHash,
		Kind:                   events.SignKindStakeImport,
		CompressedGenPubKeyHex: joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	return []*core.SignRequest{&exportTxSignReq, &importTxSignReq}, nil
}

func (c *TaskCreator) createTxs(stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*Txs, error) {
	nodeID, _ := ids.ShortFromPrefixedString(stakeReq.NodeID, ids.NodeIDPrefix)
	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(stakeReq.Amount, 10)

	startTime := big.NewInt(stakeReq.StartTime)
	endTIme := big.NewInt(stakeReq.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New(ErrMsgInvalidUint64)
	}

	cChainAddr, _ := stakeReq.GenPubKey.CChainAddress()
	pChainAddr, _ := stakeReq.GenPubKey.PChainAddress()

	st := Txs{
		ReqNo:         stakeReq.ReqNo,
		Nonce:         c.NonceGiver.GetNonce(stakeReq.ReqNo),
		ReqHash:       reqHash.String(),
		DelegateAmt:   nAVAXAmount.Uint64(),
		StartTime:     startTime.Uint64(),
		EndTime:       endTIme.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,
		NodeID:        ids.NodeID(nodeID),

		BaseFeeGwei: cchain.BaseFeeGwei,
		NetworkID:   c.Network.NetworkID(),
		CChainID:    c.Network.CChainID(),
		Asset:       c.Network.Asset(),
		ImportFee:   c.Network.ImportFee(),
	}

	return &st, nil
}
