package mpcclient

import (
	"context"
	"encoding/hex"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/avalido/mpc-controller/core/mpc"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
)

var (
	_ mpc.MpcClient = (*SimulatingClient)(nil)
)

func NewSimulatingClient(privateKey string) (*SimulatingClient, error) {
	sk, err := crypto.ParsePrivateKeySECP256K1R(privateKey)
	if err != nil {
		return nil, err
	}
	return &SimulatingClient{
		Logger:       nil,
		privateKey:   sk,
		keygenReqMap: map[string]struct{}{},
		signReqMap:   map[string]*mpc.SignRequest{},
	}, nil
}

func (c *SimulatingClient) UncompressedPublicKeyBytes() []byte {
	pk := c.privateKey.PublicKey().Bytes()
	pkEcdsa, _ := ethCrypto.DecompressPubkey(pk)
	pkb := ethCrypto.FromECDSAPub(pkEcdsa)
	return pkb
}

type SimulatingClient struct {
	Logger       logger.Logger
	privateKey   *avaCrypto.PrivateKeySECP256K1R
	keygenReqMap map[string]struct{}
	signReqMap   map[string]*mpc.SignRequest
}

func (c *SimulatingClient) Keygen(ctx context.Context, in *mpc.KeygenRequest, opts ...grpc.CallOption) (*mpc.KeygenResponse, error) {
	c.keygenReqMap[in.RequestId] = struct{}{}
	resp := &mpc.KeygenResponse{}
	resp.RequestId = in.RequestId
	return resp, nil
}

func (c *SimulatingClient) Sign(ctx context.Context, in *mpc.SignRequest, opts ...grpc.CallOption) (*mpc.SignResponse, error) {
	c.signReqMap[in.RequestId] = in
	resp := &mpc.SignResponse{}
	resp.RequestId = in.RequestId
	return resp, nil
}

func (c *SimulatingClient) CheckResult(ctx context.Context, in *mpc.CheckResultRequest, opts ...grpc.CallOption) (*mpc.CheckResultResponse, error) {
	if _, ok := c.keygenReqMap[in.RequestId]; ok {
		resp := &mpc.CheckResultResponse{}
		resp.RequestId = in.RequestId
		resp.RequestType = 1 // KEYGEN
		resp.RequestStatus = mpc.CheckResultResponse_DONE
		resp.Result = hex.EncodeToString(c.privateKey.PublicKey().Bytes())
		return resp, nil
	} else if s, ok := c.signReqMap[in.RequestId]; ok {
		hashBytes, err := hex.DecodeString(s.Hash)
		if err != nil {
			return nil, err
		}
		sig, err := c.privateKey.SignHash(hashBytes)
		if err != nil {
			return nil, err
		}
		resp := &mpc.CheckResultResponse{}
		resp.RequestId = in.RequestId
		resp.RequestType = 2 // SIGN
		resp.RequestStatus = mpc.CheckResultResponse_DONE
		resp.Result = hex.EncodeToString(sig)
		return resp, nil
	}
	return nil, nil
}
