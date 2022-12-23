package mpcclient

import (
	"context"
	"encoding/hex"
	"github.com/avalido/mpc-controller/core/mpc"
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
	client, _ := NewSimulatingClient(privKeyStr)
	kg := &mpc.KeygenRequest{}
	kg.RequestId = "kg"
	kg.ParticipantPublicKeys = []string{}
	kg.Threshold = 0
	_, err := client.Keygen(context.Background(), kg)
	require.Nil(t, err)
	c := &mpc.CheckResultRequest{}
	c.RequestId = "kg"
	res, _ := client.CheckResult(context.Background(), c)

	pkBytes, _ := hex.DecodeString(res.Result)
	pk, _ := ethCrypto.DecompressPubkey(pkBytes)

	require.Equal(t, privateKey.PublicKey, *pk)
	s := &mpc.SignRequest{}
	s.RequestId = reqIDStr
	s.Hash = hashStr
	s.ParticipantPublicKeys = []string{}
	s.PublicKey = ""
	_, err = client.Sign(context.Background(), s)
	require.Nil(t, err)
	pubKey := ethCrypto.FromECDSAPub(&privateKey.PublicKey)

	c.RequestId = reqIDStr
	res, _ = client.CheckResult(context.Background(), c)
	sig, _ := hex.DecodeString(res.Result)

	pubKeyRecovered, err := ethCrypto.Ecrecover(hash, sig)
	require.NoError(t, err)
	require.Equal(t, pubKey, pubKeyRecovered)
}
