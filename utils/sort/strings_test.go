package sort

import (
	"fmt"
	"testing"
)

func TestSortMultiLineStr(t *testing.T) {
	raw := `	Logger           logger.Logger
	ContractAddr     common.Address
	RewardUTXOGetter chain.RewardUTXOGetter
	Dispatcher       dispatcher.DispatcherClaasic
	MyPubKeyHashHex  string
	Cache            cache.ICache
	SignDoner        core.SignDoner
	Signer           *bind.TransactOpts
	Transactor       bind.ContractTransactor
	Receipter        chain.Receipter
	chain.NetworkContext
	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer`

	fmt.Println(SortMultiLineStr(raw))
}
