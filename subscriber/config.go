package subscriber

import "github.com/ethereum/go-ethereum/common"

type Config struct {
	EthWsURL          string
	MpcManagerAddress common.Address
}
