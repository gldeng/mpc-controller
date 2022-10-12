package atomicTask

import "github.com/ava-labs/avalanchego/ids"

var _ AtomicTx = (*StakeAtomicTx)(nil)

type StakeAtomicTx struct{}

func (t *StakeAtomicTx) RequestID() string {
	return ""
}

func (t *StakeAtomicTx) SourceChain() SourceChain {
	return SourceChainCChain
}

func (t *StakeAtomicTx) ExportTxHash() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetExportTxSig(sig [65]byte) error {
	return nil
}

func (t *StakeAtomicTx) SignedExportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetExportTxID(id ids.ID) {

}

func (t *StakeAtomicTx) ImportTxHash() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetImportTxSig(sig [65]byte) error {
	return nil
}

func (t *StakeAtomicTx) SignedImportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetImportTxID(id ids.ID) {

}

func (t *StakeAtomicTx) String() string {
	return ""
}
