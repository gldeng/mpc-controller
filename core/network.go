package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
)

type NetworkContext struct {
	networkID  uint32
	cChainID   ids.ID
	asset      avax.Asset
	importFee  uint64
	gasPerByte uint64
	gasPerSig  uint64
	gasFixed   uint64
}

func NewNetworkContext(networkID uint32,
	cChainID ids.ID,
	asset avax.Asset,
	importFee uint64,
	gasPerByte uint64,
	gasPerSig uint64,
	gasFixed uint64) NetworkContext {
	return NetworkContext{
		networkID:  networkID,
		cChainID:   cChainID,
		asset:      asset,
		importFee:  importFee,
		gasPerByte: gasPerByte,
		gasPerSig:  gasPerSig,
		gasFixed:   gasFixed,
	}
}

func (c *NetworkContext) NetworkID() uint32 {
	return c.networkID
}

func (c *NetworkContext) CChainID() ids.ID {
	return c.cChainID
}

func (c *NetworkContext) Asset() avax.Asset {
	return c.asset
}

func (c *NetworkContext) ImportFee() uint64 {
	return c.importFee
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
