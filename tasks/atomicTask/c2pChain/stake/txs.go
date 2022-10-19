package stake

import (
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	SigLength = 65
)

type Txs struct {
	ReqNo         uint64
	Nonce         uint64
	ReqHash       string
	DelegateAmt   uint64
	StartTime     uint64
	EndTime       uint64
	CChainAddress common.Address
	PChainAddress ids.ShortID
	NodeID        ids.NodeID

	BaseFeeGwei uint64
	NetworkID   uint32
	CChainID    ids.ID
	Asset       avax.Asset
	ImportFee   uint64

	exportTx       *evm.UnsignedExportTx
	exportTxCred   *secp256k1fx.Credential
	signedExportTx *evm.Tx

	importTx       *txs.ImportTx
	importTxCred   *secp256k1fx.Credential
	signedImportTx *txs.Tx
}

// ---

func (t *Txs) ExportTxHash() ([]byte, error) {
	exportTx, err := t.buildUnsignedExportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tx := evm.Tx{
		UnsignedAtomicTx: exportTx,
	}
	unsignedBytes, err := evm.Codec.Marshal(txs.Version, &tx.UnsignedAtomicTx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func (t *Txs) SetExportTxSig(sig [SigLength]byte) error {
	if t.exportTxCred != nil {
		return errors.New(ErrMsgSignatureAlreadySet)
	}
	hash, err := t.ExportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}
	cred, err := t.validateAndGetCred(hash, sig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.exportTxCred = cred
	return nil
}

func (t *Txs) SignedExportTxBytes() ([]byte, error) {
	tx, err := t.getSignedExportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t.signedExportTx = tx

	return tx.SignedBytes(), nil
}

func (t *Txs) ExportTxID() ids.ID {
	return t.signedExportTx.ID()
}

func (t *Txs) ImportTxHash() ([]byte, error) {
	importTx, err := t.buildUnsignedImportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tx := txs.Tx{
		Unsigned: importTx,
	}
	unsignedBytes, err := platformvm.Codec.Marshal(txs.Version, &tx.Unsigned)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func (t *Txs) SetImportTxSig(sig [SigLength]byte) error {
	if t.importTxCred != nil {
		return errors.New(ErrMsgSignatureAlreadySet)
	}
	hash, err := t.ImportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}
	cred, err := t.validateAndGetCred(hash, sig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.importTxCred = cred
	return nil
}

func (t *Txs) SignedImportTxBytes() ([]byte, error) {
	tx, err := t.getSignedImportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	t.signedImportTx = tx
	return tx.Bytes(), nil
}

func (t *Txs) SingedImportTxUTXOs() ([]*avax.UTXO, error) {
	signedImportTx, err := t.getSignedImportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return signedImportTx.UTXOs(), nil
}

func (t *Txs) ImportTxID() ids.ID {
	return t.signedImportTx.ID()
}

func (t *Txs) String() string {
	// todo:
	return ""
}

// ---

func (t *Txs) buildUnsignedExportTx() (*evm.UnsignedExportTx, error) {
	if t.exportTx != nil {
		return t.exportTx, nil
	}
	exportAmt := t.DelegateAmt + t.ImportFee
	input := evm.EVMInput{
		Address: t.CChainAddress,
		Amount:  exportAmt,
		AssetID: t.Asset.ID,
		Nonce:   t.Nonce,
	}
	var outs []*avax.TransferableOutput
	outs = append(outs, &avax.TransferableOutput{
		Asset: t.Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: exportAmt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					t.PChainAddress,
				},
			},
		},
	})

	tx := &evm.UnsignedExportTx{
		NetworkID:        t.NetworkID,
		BlockchainID:     t.CChainID,
		DestinationChain: ids.Empty,
		Ins: []evm.EVMInput{
			input,
		},
		ExportedOutputs: outs,
	}

	gas, err := tx.GasUsed(true)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	exportFee := gas * t.BaseFeeGwei
	tx.Ins[0].Amount += exportFee

	t.exportTx = tx
	return t.exportTx, nil
}

func (t *Txs) getSignedExportTx() (*evm.Tx, error) {
	unsignedTx, err := t.buildUnsignedExportTx()
	if err != nil {
		return nil, errors.Wrapf(err, ErrMsgMissingCredential)
	}

	if t.exportTxCred == nil {
		return nil, errors.New(ErrMsgMissingCredential)
	}

	tx := evm.Tx{
		UnsignedAtomicTx: unsignedTx,
		Creds: []verify.Verifiable{
			t.exportTxCred,
		},
	}
	err = tx.Sign(evm.Codec, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &tx, nil
}

// ---

func (t *Txs) buildUnsignedImportTx() (*txs.ImportTx, error) {
	if t.importTx != nil {
		return t.importTx, nil
	}
	signedExportTx, err := t.getSignedExportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	exportTx := signedExportTx.UnsignedAtomicTx.(*evm.UnsignedExportTx)
	index := uint32(0)
	amt := exportTx.ExportedOutputs[index].Out.Amount()
	utxo := t.paySelf(amt - t.ImportFee)
	t.importTx = &txs.ImportTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    t.NetworkID,
			BlockchainID: ids.Empty,
			Outs: []*avax.TransferableOutput{
				utxo,
			},
		}},
		SourceChain: t.CChainID,
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        exportTx.ID(),
				OutputIndex: index,
			},
			Asset: t.Asset,
			In: &secp256k1fx.TransferInput{
				Amt: amt,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
		}},
	}
	return t.importTx, nil
}

func (t *Txs) paySelf(amt uint64) *avax.TransferableOutput {
	return &avax.TransferableOutput{
		Asset: t.Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: amt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					t.PChainAddress,
				},
			},
		},
	}
}

func (t *Txs) getSignedImportTx() (*txs.Tx, error) {
	unsignedTx, err := t.buildUnsignedImportTx()
	if err != nil {
		return nil, errors.Wrapf(err, ErrMsgMissingCredential)
	}

	if t.importTxCred == nil {
		return nil, errors.New(ErrMsgMissingCredential)
	}

	tx := txs.Tx{
		Unsigned: unsignedTx,
		Creds: []verify.Verifiable{
			t.importTxCred,
		},
	}
	err = tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &tx, nil
}

// ---

func (t *Txs) validateAndGetCred(hash []byte, sig [SigLength]byte) (*secp256k1fx.Credential, error) {
	sigIndex := 0
	numSigs := 1
	cred := &secp256k1fx.Credential{
		Sigs: make([][SigLength]byte, numSigs),
	}
	copy(cred.Sigs[sigIndex][:], sig[:])

	pk, err := new(avaCrypto.FactorySECP256K1R).RecoverHashPublicKey(hash, sig[:])
	if err != nil {
		return nil, errors.New(ErrMsgPubKeyCannotRecover)
	}
	if t.PChainAddress != pk.Address() {
		return nil, errors.New(ErrMsgSignatureInvalid)
	}
	return cred, nil
}
