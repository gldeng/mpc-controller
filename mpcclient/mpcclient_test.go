package mpcclient

import (
	"context"
	"encoding/hex"
	"github.com/avalido/mpc-controller/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimulatingMpcClient(t *testing.T) {
	reqIDStr := "123"
	privKeyStr := "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
	hashStr := "1111111111111111111111111111111111111111111111111111111111111111"
	hash, _ := hex.DecodeString(hashStr)
	privateKey, _ := ethCrypto.HexToECDSA(privKeyStr)
	client, _ := NewSimulatingMpcClient(privKeyStr)
	client.Keygen(context.Background(), &types.KeygenRequest{
		ReqID:                  "kg",
		CompressedPartiPubKeys: []string{},
		Threshold:              0,
	})
	res, _ := client.Result(context.Background(), "kg")

	pkBytes, _ := hex.DecodeString(res.Result)
	pk, _ := ethCrypto.DecompressPubkey(pkBytes)

	require.Equal(t, privateKey.PublicKey, *pk)
	client.Sign(context.Background(), &types.SignRequest{
		ReqID:                  reqIDStr,
		Hash:                   hashStr,
		CompressedGenPubKeyHex: "",
		CompressedPartiPubKeys: []string{},
	})
	pubKey := ethCrypto.FromECDSAPub(&privateKey.PublicKey)

	res, _ = client.Result(context.Background(), "123")
	sig, _ := hex.DecodeString(res.Result)

	pubKeyRecovered, err := ethCrypto.Ecrecover(hash, sig)
	require.NoError(t, err)
	require.Equal(t, pubKey, pubKeyRecovered)
}
