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
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	SigLength = 65 // todo: consider reuse
)

type Tx struct {
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

	UTXOsToPay []*avax.UTXO

	addDelegatorTx     *txs.AddDelegatorTx
	addDelegatorTxCred *secp256k1fx.Credential
	addDelegatorTxID   ids.ID
}

// ---

func (t *Tx) AddDelegatorTxHash() ([]byte, error) {
	unsignedTx, err := t.buildUnsignedAddDelegatorTx()
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

func (t *Tx) SetAddDelegatorTxSig(sig [SigLength]byte) error {
	if t.addDelegatorTxCred != nil {
		return errors.New(ErrMsgSignatureAlreadySet)
	}

	hash, err := t.AddDelegatorTxHash()
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

func (t *Tx) SignedAddDelegatorTxBytes() ([]byte, error) {
	tx, err := t.getSignedAddDelegatorTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tx.Bytes(), nil
}

func (t *Tx) SetAddDelegatorTxID(id ids.ID) {
	t.addDelegatorTxID = id
}

// ---

func (t *Tx) buildUnsignedAddDelegatorTx() (*txs.AddDelegatorTx, error) {
	if t.addDelegatorTx != nil {
		return t.addDelegatorTx, nil
	}

	var (
		signersPlaceholder []*avaCrypto.PrivateKeySECP256K1R
		ins                []*avax.TransferableInput
		stakedOuts         []*avax.TransferableOutput
	)

	utxo := t.UTXOsToPay[0]

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

func (t *Tx) getSignedAddDelegatorTx() (*txs.Tx, error) {
	unsignedTx, err := t.buildUnsignedAddDelegatorTx()

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

func (t *Tx) paySelf(amt uint64) *avax.TransferableOutput { // todo: reuse across packages
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

func (t *Tx) validateAndGetCred(hash []byte, sig [SigLength]byte) (*secp256k1fx.Credential, error) { // todo: reuse across packages
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
