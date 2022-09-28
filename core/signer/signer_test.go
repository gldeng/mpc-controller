// todo: add more test cases

package signer

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/tasks/staking/mocks"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
)

type SignRequesterTestSuite struct {
	suite.Suite
	SignRequestArgs
}

func (suite *SignRequesterTestSuite) SetupTest() {
	suite.SignRequestArgs = SignRequestArgs{
		TaskID:                 "0x74045c1e4c9c538ddfede3971ba8df6fb9a0de096459ab7f5152334c354ba060",
		CompressedPartiPubKeys: []string{"03d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e", "0373ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6"},
		CompressedGenPubKeyHex: "0373ee8ae85eceeb6547a31ad6a5add1598f35964e1e3f0513bcde455f4c15b2de",
	}
}

func (suite *SignRequesterTestSuite) TestSignExportTx() {
	require := suite.Require()

	ctx := context.Background()
	hash := "3f92d8a04c3f66c43fca816ed18296adcbc2d35b181945af9a9c28517b4e6e6d"
	signature := "c3ec6ebfc1c44719550c138795e7a69e356e84be17f6ef968523bf0ebb30bdbe7bc99b420d6d5f070056e3495fa3c6e51d9195400a322b1fa061a90a517232a400"

	exportTxSignReq := core.SignRequest{
		ReqID:                  suite.TaskID + "-" + strconv.Itoa(0),
		CompressedGenPubKeyHex: suite.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: suite.CompressedPartiPubKeys,
		Hash:                   hash,
	}
	mockResultFn := func(ctx context.Context, request *core.SignRequest) *core.Result {
		output := &core.Result{
			ReqID:     request.ReqID,
			Result:    signature,
			ReqType:   "SIGN",
			ReqStatus: "DONE",
		}

		return output
	}

	signDoner := &mocks.SignDoner{} // todo: use NewSignDoner() and Expecter Interfaces.
	signDoner.On("SignDone", ctx, &exportTxSignReq).Return(mockResultFn, nil)

	signRequester := &Signer{
		SignDoner:       signDoner,
		SignRequestArgs: suite.SignRequestArgs,
	}

	sigBytes, err := signRequester.SignExportTx(ctx, bytes.HexToBytes(hash))
	require.Nil(err)
	require.Equal(signature, bytes.Bytes65ToHex(sigBytes))
}

func (suite *SignRequesterTestSuite) TestSignImportTx() {
	require := suite.Require()

	ctx := context.Background()
	hash := "98d7f7f48cd1accc38b78471db8865b4c830fff71713413e5c12c51bcbd8c388"
	signature := "e3a7a5130e7848c51d9b64abb21a90d96cf08e5192a93ee95d49a6c2faf876fd59570d255db00d513bf693153c73d1d4b169d2f290e676ead48aaf9269cad72700"

	exportTxSignReq := core.SignRequest{
		ReqID:                  suite.TaskID + "-" + strconv.Itoa(1),
		CompressedGenPubKeyHex: suite.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: suite.CompressedPartiPubKeys,
		Hash:                   hash,
	}
	mockResultFn := func(ctx context.Context, request *core.SignRequest) *core.Result {
		output := &core.Result{
			ReqID:     request.ReqID,
			Result:    signature,
			ReqType:   "SIGN",
			ReqStatus: "DONE",
		}

		return output
	}

	signDoner := &mocks.SignDoner{}
	signDoner.On("SignDone", ctx, &exportTxSignReq).Return(mockResultFn, nil)

	signRequester := &Signer{
		SignDoner:       signDoner,
		SignRequestArgs: suite.SignRequestArgs,
	}

	sigBytes, err := signRequester.SignImportTx(ctx, bytes.HexToBytes(hash))
	require.Nil(err)
	require.Equal(signature, bytes.Bytes65ToHex(sigBytes))
}

func (suite *SignRequesterTestSuite) TestSignAddDelegatorTx() {
	require := suite.Require()

	ctx := context.Background()
	hash := "dbf9dfcd082e4f6b1937c1dc608d693041ee8a7ee76d3446c7ff957de8696fda"
	signature := "94059a6d9f4ae471169369755012050c37884d36b3a83b2b79b059c90905f7ba2372ca48eecc9b48f46155b1c4c16891365f14c0e830076502f704f1f38a16da00"

	exportTxSignReq := core.SignRequest{
		ReqID:                  suite.TaskID + "-" + strconv.Itoa(2),
		CompressedGenPubKeyHex: suite.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: suite.CompressedPartiPubKeys,
		Hash:                   hash,
	}
	mockResultFn := func(ctx context.Context, request *core.SignRequest) *core.Result {
		output := &core.Result{
			ReqID:     request.ReqID,
			Result:    signature,
			ReqType:   "SIGN",
			ReqStatus: "DONE",
		}

		return output
	}

	signDoner := &mocks.SignDoner{}
	signDoner.On("SignDone", ctx, &exportTxSignReq).Return(mockResultFn, nil)

	signRequester := &Signer{
		SignDoner:       signDoner,
		SignRequestArgs: suite.SignRequestArgs,
	}

	sigBytes, err := signRequester.SignAddDelegatorTx(ctx, bytes.HexToBytes(hash))
	require.Nil(err)
	require.Equal(signature, bytes.Bytes65ToHex(sigBytes))
}

func TestSignRequesterTestSuite(t *testing.T) {
	suite.Run(t, new(SignRequesterTestSuite))
}
