package staking

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/pkg/errors"
	"math/big"
)

type StakeTaskCreator struct {
	NetWorkCtx chain.NetworkContext
	NonceGiver noncer.Noncer
}

func (eh *StakeTaskCreator) CreateStakeTask(stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*StakeTask, error) {
	nodeID, _ := ids.ShortFromPrefixedString(stakeReq.NodeID, ids.NodeIDPrefix)
	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(stakeReq.Amount, 10)

	startTime := big.NewInt(stakeReq.StartTime)
	endTIme := big.NewInt(stakeReq.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New("invalid uint64")
	}

	cChainAddr, _ := stakeReq.GenPubKey.CChainAddress()
	pChainAddr, _ := stakeReq.GenPubKey.PChainAddress()

	st := StakeTask{
		ReqNo:         stakeReq.ReqNo,
		Nonce:         eh.NonceGiver.GetNonce(stakeReq.ReqNo),
		ReqHash:       reqHash.String(),
		DelegateAmt:   nAVAXAmount.Uint64(),
		StartTime:     startTime.Uint64(),
		EndTime:       endTIme.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,
		NodeID:        ids.NodeID(nodeID),
		BaseFeeGwei:   cchain.BaseFeeGwei,
		Network:       eh.NetWorkCtx,
	}

	return &st, nil
}
