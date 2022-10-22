package chain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"math/big"
)

// todo: use exported fields for convinience

type NetworkContext struct {
	chainID     *big.Int
	networkID   uint32
	cChainID    ids.ID
	asset       avax.Asset
	importFee   uint64
	exportFee   uint64
	gasPerByte  uint64
	gasPerSig   uint64
	gasFixed    uint64
	baseFeeGwei uint64
}

func NewNetworkContext(networkID uint32,
	cChainID ids.ID,
	chainID *big.Int,
	asset avax.Asset,
	importFee,
	exportFee,
	gasPerByte,
	gasPerSig,
	gasFixed,
	baseFeeGwei uint64,
) NetworkContext {
	return NetworkContext{
		chainID:     chainID,
		networkID:   networkID,
		cChainID:    cChainID,
		asset:       asset,
		importFee:   importFee,
		exportFee:   exportFee,
		gasPerByte:  gasPerByte,
		gasPerSig:   gasPerSig,
		gasFixed:    gasFixed,
		baseFeeGwei: baseFeeGwei,
	}
}

func (c *NetworkContext) NetworkID() uint32 {
	return c.networkID
}

func (c *NetworkContext) CChainID() ids.ID {
	return c.cChainID
}

func (c *NetworkContext) ChainID() *big.Int {
	return c.chainID
}

func (c *NetworkContext) Asset() avax.Asset {
	return c.asset
}

func (c *NetworkContext) ImportFee() uint64 {
	return c.importFee
}

func (c *NetworkContext) ExportFee() uint64 {
	return c.exportFee
}

func (c *NetworkContext) GasPerByte() uint64 {
	return c.gasPerByte
}

func (c *NetworkContext) GasPerSig() uint64 {
	return c.gasPerSig
}

func (c *NetworkContext) GasFixed() uint64 {
	return c.gasFixed
}

func (c *NetworkContext) BaseFeeGwei() uint64 {
	return c.baseFeeGwei
}
