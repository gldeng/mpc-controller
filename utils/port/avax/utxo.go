package avax

import (
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

type UTXO struct {
	avax.UTXOID
	avax.Asset
	Out *secp256k1fx.TransferOutput
}
