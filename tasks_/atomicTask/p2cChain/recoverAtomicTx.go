package p2cChain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/tasks_/atomicTask/c2pChain/stake"
)

var _ stake.AtomicTx = (*RecoverAtomicTx)(nil)

type RecoverAtomicTx struct{}

func (t *RecoverAtomicTx) RequestID() string {
	return ""
}

func (t *RecoverAtomicTx) SourceChain() stake.SourceChain {
	return stake.SourceChainCChain
}

func (t *RecoverAtomicTx) ExportTxHash() ([]byte, error) {
	return nil, nil
}

func (t *RecoverAtomicTx) SetExportTxSig(sig [stake.SigLength]byte) error {
	return nil
}

func (t *RecoverAtomicTx) SignedExportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *RecoverAtomicTx) SetExportTxID(id ids.ID) {

}

func (t *RecoverAtomicTx) ImportTxHash() ([]byte, error) {
	return nil, nil
}

func (t *RecoverAtomicTx) SetImportTxSig(sig [stake.SigLength]byte) error {
	return nil
}

func (t *RecoverAtomicTx) SignedImportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *RecoverAtomicTx) SetImportTxID(id ids.ID) {

}

func (t *RecoverAtomicTx) String() string {
	return ""
}
