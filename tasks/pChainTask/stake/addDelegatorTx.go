package stake

import (
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/events"
	"github.com/pkg/errors"
)

const (
	SigLength = 65 // todo: consider reuse
)

type AddDelegatorTx struct {
	*events.StakeAtomicTaskHandled

	NetworkID uint32
	Asset     avax.Asset

	addDelegatorTx       *txs.AddDelegatorTx
	addDelegatorTxCred   *secp256k1fx.Credential
	signedAddDelegatorTx *txs.Tx
}

// ---

func (t *AddDelegatorTx) TxHash() ([]byte, error) {
	unsignedTx, err := t.buildUnsignedTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tx := txs.Tx{
		Unsigned: unsignedTx,
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

func (t *AddDelegatorTx) SetTxSig(sig [SigLength]byte) error {
	if t.addDelegatorTxCred != nil {
		return errors.New(ErrMsgSignatureAlreadySet)
	}

	hash, err := t.TxHash()
	if err != nil {
		return errors.WithStack(err)
	}
	cred, err := t.validateAndGetCred(hash, sig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.addDelegatorTxCred = cred
	return nil
}

func (t *AddDelegatorTx) SignedTxBytes() ([]byte, error) {
	tx, err := t.getSignedTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t.signedAddDelegatorTx = tx
	return tx.Bytes(), nil
}

func (t *AddDelegatorTx) ID() ids.ID {
	return t.signedAddDelegatorTx.ID()
}

func (t *AddDelegatorTx) String() string {
	// todo:
	return ""
}

// ---

func (t *AddDelegatorTx) buildUnsignedTx() (*txs.AddDelegatorTx, error) {
	if t.addDelegatorTx != nil {
		return t.addDelegatorTx, nil
	}

	var (
		signersPlaceholder []*avaCrypto.PrivateKeySECP256K1R
		ins                []*avax.TransferableInput
		stakedOuts         []*avax.TransferableOutput
	)

	utxo := t.UTXOsToStake[0]

	stakedOuts = append(stakedOuts, t.paySelf(utxo.Out.(*secp256k1fx.TransferOutput).Amt))
	ins = append(ins, spend(utxo))
	signers := [][]*avaCrypto.PrivateKeySECP256K1R{
		signersPlaceholder,
	}
	avax.SortTransferableInputsWithSigners(ins, signers)
	t.addDelegatorTx = &txs.AddDelegatorTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    t.NetworkID,
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

func (t *AddDelegatorTx) getSignedTx() (*txs.Tx, error) {
	unsignedTx, err := t.buildUnsignedTx()

	if err != nil {
		return nil, errors.Wrapf(err, ErrMsgMissingCredential)
	}

	if t.addDelegatorTxCred == nil {
		return nil, errors.New(ErrMsgMissingCredential)
	}

	tx := txs.Tx{
		Unsigned: unsignedTx,
		Creds: []verify.Verifiable{
			t.addDelegatorTxCred,
		},
	}
	err = tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &tx, nil
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

// ---

func (t *AddDelegatorTx) paySelf(amt uint64) *avax.TransferableOutput { // todo: reuse across packages
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

func (t *AddDelegatorTx) validateAndGetCred(hash []byte, sig [SigLength]byte) (*secp256k1fx.Credential, error) { // todo: reuse across packages
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
