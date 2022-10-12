package atomicTask

import "github.com/ava-labs/avalanchego/ids"

var _ AtomicTx = (*RecoverAtomicTx)(nil)

type RecoverAtomicTx struct{}

func (t *RecoverAtomicTx) RequestID() string {
	return ""
}

func (t *RecoverAtomicTx) SourceChain() SourceChain {
	return SourceChainCChain
}

func (t *RecoverAtomicTx) ExportTxHash() ([]byte, error) {
	return nil, nil
}

func (t *RecoverAtomicTx) SetExportTxSig(sig [65]byte) error {
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

func (t *RecoverAtomicTx) SetImportTxSig(sig [65]byte) error {
	return nil
}

func (t *RecoverAtomicTx) SignedImportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *RecoverAtomicTx) SetImportTxID(id ids.ID) {

}
