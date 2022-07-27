package staking

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

const (
	// These should be imported from package. But currently they are not public. Duplicate it here for the time being.
	evmCodecVersion        = uint16(0)
	platformvmCodecVersion = uint16(0)
	sigLength              = 65
)

type StakeTask struct {
	RequestID          uint64
	Nonce              uint64
	TaskID             string
	network            chain.NetworkContext
	DelegateAmt        uint64
	StartTime          uint64
	EndTime            uint64
	baseFeeGwei        uint64
	CChainAddress      common.Address
	PChainAddress      ids.ShortID
	NodeID             ids.NodeID
	exportTx           *evm.UnsignedExportTx
	importTx           *txs.ImportTx
	addDelegatorTx     *txs.AddDelegatorTx
	exportTxCred       *secp256k1fx.Credential
	importTxCred       *secp256k1fx.Credential
	addDelegatorTxCred *secp256k1fx.Credential
	factory            avaCrypto.FactorySECP256K1R
}

func NewStakeTask(taskID string, networkContext chain.NetworkContext, reqId uint64, pubkey ecdsa.PublicKey, nonce uint64, nodeID ids.NodeID, delegateAmt uint64,
	startTime uint64, endTime uint64,
	baseFeeGwei uint64) (*StakeTask, error) {
	addr, err := ids.ToShortID(hashing.PubkeyBytesToAddress(serializeCompresed(&pubkey)))
	if err != nil {
		return nil, err
	}
	return &StakeTask{
		TaskID:        taskID,
		network:       networkContext,
		Nonce:         nonce,
		DelegateAmt:   delegateAmt,
		baseFeeGwei:   baseFeeGwei,
		StartTime:     startTime,
		EndTime:       endTime,
		CChainAddress: crypto.PubkeyToAddress(pubkey),
		PChainAddress: addr,
		RequestID:     reqId,
		NodeID:        nodeID,
		factory:       avaCrypto.FactorySECP256K1R{},
	}, nil
}

func (t *StakeTask) ExportTxHash() ([]byte, error) {
	exportTx, err := t.buildUnsignedExportTx()
	if err != nil {
		return nil, err
	}
	tx := evm.Tx{
		UnsignedAtomicTx: exportTx,
	}
	unsignedBytes, err := evm.Codec.Marshal(evmCodecVersion, &tx.UnsignedAtomicTx)
	if err != nil {
		return nil, err
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *StakeTask) ImportTxHash() ([]byte, error) {
	importTx, err := t.buildUnsignedImportTx()
	if err != nil {
		return nil, err
	}
	tx := txs.Tx{
		Unsigned: importTx,
	}
	unsignedBytes, err := platformvm.Codec.Marshal(platformvmCodecVersion, &tx.Unsigned)
	if err != nil {
		return nil, err
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *StakeTask) AddDelegatorTxHash() ([]byte, error) {
	unsignedTx, err := t.buildUnsignedAddDelegatorTx()
	if err != nil {
		return nil, err
	}
	tx := txs.Tx{
		Unsigned: unsignedTx,
	}
	unsignedBytes, err := platformvm.Codec.Marshal(platformvmCodecVersion, &tx.Unsigned)
	if err != nil {
		return nil, err
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *StakeTask) SetExportTxSig(sig [sigLength]byte) error {
	if t.exportTxCred != nil {
		return errors.New("exportTxSig already set")
	}
	hash, err := t.ExportTxHash()
	if err != nil {
		return err
	}
	cred, err := t.validateAndGetCred(hash, sig)
	if err != nil {
		return err
	}
	t.exportTxCred = cred
	return nil
}

func (t *StakeTask) SetImportTxSig(sig [sigLength]byte) error {
	if t.importTxCred != nil {
		return errors.New("importTxSig already set")
	}
	hash, err := t.ImportTxHash()
	if err != nil {
		return err
	}
	cred, err := t.validateAndGetCred(hash, sig)
	if err != nil {
		return err
	}
	t.importTxCred = cred
	return nil
}

func (t *StakeTask) SetAddDelegatorTxSig(sig [sigLength]byte) error {
	if t.addDelegatorTxCred != nil {
		return errors.New("addDelegatorTxSig already set")
	}

	hash, err := t.AddDelegatorTxHash()
	if err != nil {
		return err
	}
	cred, err := t.validateAndGetCred(hash, sig)
	if err != nil {
		return err
	}
	t.addDelegatorTxCred = cred
	return nil
}

func (t *StakeTask) GetSignedExportTx() (*evm.Tx, error) {
	unsignedTx, err := t.buildUnsignedExportTx()
	if err != nil {
		return nil, errors.New("missing ExportTx cred")
	}

	if t.exportTxCred == nil {
		return nil, errors.New("missing ExportTx cred")
	}

	tx := evm.Tx{
		UnsignedAtomicTx: unsignedTx,
		Creds: []verify.Verifiable{
			t.exportTxCred,
		},
	}
	err = tx.Sign(evm.Codec, nil)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (t *StakeTask) GetSignedImportTx() (*txs.Tx, error) {
	var noTxCredErr = errors.New("missing ImportTx cred")
	unsignedTx, err := t.buildUnsignedImportTx()

	if err != nil {
		return nil, err
	}

	if t.importTxCred == nil {
		return nil, noTxCredErr
	}

	tx := txs.Tx{
		Unsigned: unsignedTx,
		Creds: []verify.Verifiable{
			t.importTxCred,
		},
	}
	err = tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, err
	}
	return &tx, nil

}

func (t *StakeTask) GetSignedAddDelegatorTx() (*txs.Tx, error) {
	var noTxCredErr = errors.New("no tx cred")
	unsignedTx, err := t.buildUnsignedAddDelegatorTx()

	if err != nil {
		return nil, err
	}

	if t.addDelegatorTxCred == nil {
		return nil, noTxCredErr
	}

	tx := txs.Tx{
		Unsigned: unsignedTx,
		Creds: []verify.Verifiable{
			t.addDelegatorTxCred,
		},
	}
	err = tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, err
	}
	return &tx, nil

}

func (t *StakeTask) buildUnsignedExportTx() (*evm.UnsignedExportTx, error) {
	if t.exportTx != nil {
		return t.exportTx, nil
	}
	exportAmt := t.DelegateAmt + t.network.ImportFee()
	input := evm.EVMInput{
		Address: t.CChainAddress,
		Amount:  exportAmt,
		AssetID: t.network.Asset().ID,
		Nonce:   t.Nonce,
	}
	var outs []*avax.TransferableOutput
	outs = append(outs, &avax.TransferableOutput{
		Asset: t.network.Asset(),
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
		NetworkID:        t.network.NetworkID(),
		BlockchainID:     t.network.CChainID(),
		DestinationChain: ids.Empty,
		Ins: []evm.EVMInput{
			input,
		},
		ExportedOutputs: outs,
	}

	gas, err := tx.GasUsed(true)
	if err != nil {
		return nil, err
	}
	exportFee := gas * t.baseFeeGwei
	tx.Ins[0].Amount += exportFee

	t.exportTx = tx
	return t.exportTx, nil
}

func (t *StakeTask) buildUnsignedImportTx() (*txs.ImportTx, error) {
	if t.importTx != nil {
		return t.importTx, nil
	}
	signedExportTx, err := t.GetSignedExportTx()
	if err != nil {
		return nil, err
	}
	exportTx := signedExportTx.UnsignedAtomicTx.(*evm.UnsignedExportTx)
	index := uint32(0)
	amt := exportTx.ExportedOutputs[index].Out.Amount()
	utxo := t.paySelf(amt - t.network.ImportFee())
	t.importTx = &txs.ImportTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    t.network.NetworkID(),
			BlockchainID: ids.Empty,
			Outs: []*avax.TransferableOutput{
				utxo,
			},
		}},
		SourceChain: t.network.CChainID(),
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        exportTx.ID(),
				OutputIndex: index,
			},
			Asset: t.network.Asset(),
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

func (t *StakeTask) buildUnsignedAddDelegatorTx() (*txs.AddDelegatorTx, error) {

	if t.addDelegatorTx != nil {
		return t.addDelegatorTx, nil
	}

	signedImportTx, err := t.GetSignedImportTx()
	if err != nil {
		return nil, err
	}
	var (
		signersPlaceholder []*avaCrypto.PrivateKeySECP256K1R
		ins                []*avax.TransferableInput
		stakedOuts         []*avax.TransferableOutput
	)

	utxos := signedImportTx.UTXOs()
	utxo := utxos[0]

	stakedOuts = append(stakedOuts, t.paySelf(utxo.Out.(*secp256k1fx.TransferOutput).Amt))
	ins = append(ins, spend(utxo))
	signers := [][]*avaCrypto.PrivateKeySECP256K1R{
		signersPlaceholder,
	}
	avax.SortTransferableInputsWithSigners(ins, signers)
	t.addDelegatorTx = &txs.AddDelegatorTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    t.network.NetworkID(),
			BlockchainID: ids.Empty,
			Ins:          ins,
		}},
		Validator: validator.Validator{
			NodeID: t.NodeID,
			Start:  t.StartTime,
			End:    t.EndTime,
			Wght:   utxo.Out.(*secp256k1fx.TransferOutput).Amt,
		},
		Stake:        stakedOuts,
		RewardsOwner: &utxo.Out.(*secp256k1fx.TransferOutput).OutputOwners,
	}

	return t.addDelegatorTx, nil
}

func (t *StakeTask) paySelf(amt uint64) *avax.TransferableOutput {
	return &avax.TransferableOutput{
		Asset: t.network.Asset(),
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

func (t *StakeTask) getGas(tx *evm.UnsignedExportTx) uint64 {
	return t.network.GasFixed() + uint64(len(tx.Ins))*t.network.GasPerSig() + uint64(len(tx.Bytes()))*t.network.GasPerByte()
}

func spend(utxo *avax.UTXO) *avax.TransferableInput {
	return &avax.TransferableInput{
		UTXOID: utxo.UTXOID,
		Asset:  utxo.Asset,
		In: &secp256k1fx.TransferInput{
			Amt: utxo.Out.(*secp256k1fx.TransferOutput).Amt,
			Input: secp256k1fx.Input{
				SigIndices: []uint32{0},
			},
		},
	}
}

func serializeCompresed(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.MarshalCompressed(crypto.S256(), pub.X, pub.Y)
}

func (t *StakeTask) validateAndGetCred(hash []byte, sig [sigLength]byte) (*secp256k1fx.Credential, error) {
	sigIndex := 0
	numSigs := 1
	cred := &secp256k1fx.Credential{
		Sigs: make([][sigLength]byte, numSigs),
	}
	copy(cred.Sigs[sigIndex][:], sig[:])

	pk, err := t.factory.RecoverHashPublicKey(hash, sig[:])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to recover public key with hash %q and signature %q", bytes.BytesToHex(hash), bytes.Bytes65ToHex(sig))
	}
	if t.PChainAddress != pk.Address() {
		return nil, errors.Errorf("expected P-Chain address is %q, but got %q", t.PChainAddress, pk.Address())
	}
	return cred, nil
}
