package staking

import (
	ctlPk "github.com/avalido/mpc-controller"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"time"
)

type OnStakeRequestStartedTask struct {
	PubKeyHashHex string
	Signer        *bind.TransactOpts
	CheckRcptDur  time.Duration // duration to check transaction to see whether it is successful
	ctlPk.Manager
}
