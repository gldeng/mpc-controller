// todo: add more test cases

package porter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/utils/bytes"
	mocks2 "github.com/avalido/mpc-controller/utils/port/porter/mocks"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TokenPorterTestSuite struct {
	suite.Suite
}

func (suite *TokenPorterTestSuite) SetupTest() {}

func (suite *TokenPorterTestSuite) TestSignAndIssueTxs() {
	require := suite.Require()

	ctx := context.Background()

	txs := mocks2.NewTxs(suite.T())
	txSigner := mocks2.NewTxSigner(suite.T())
	txIssuer := mocks2.NewTxIssuer(suite.T())
	sigVerifier := mocks2.NewSigVerifier(suite.T())

	exportTxHash := bytes.HexToBytes("3273f531ba059c12f98b4cf7890608c66da392b8d5fc218d6d32041c76fdb674")
	exportTxSig := bytes.HexTo65Bytes("5b12ef4bf066a0d341f1bc4c47f597829a22f7b78dabbe1445a84b13053a2f334d7409466b993916dc0e4285911f21111460ea58da98ebe8fbb752bda74d77f301")
	txs.EXPECT().ExportTxHash().Return(exportTxHash, nil)
	txs.EXPECT().SetExportTxSig(exportTxSig).Return(nil)
	txs.EXPECT().SignedExportTxBytes().Return(exportTxHash, nil) // note: here is fake tx bytes, in production it should be generated from a true tx

	importTxHash := bytes.HexToBytes("6c99c91ed157e0c2d7cbd790ed89561b7e4bd71f550d4f176a85122afa90c135")
	importTxSig := bytes.HexTo65Bytes("db5abfea3856cdb6013adae8b8ecd006146158b9d90a31f8aa9ea21b9a08274c71a8f52078236b35a76bc5b62686ae01bfdf585e1785d7be891d6e00d70fc58000")
	txs.EXPECT().ImportTxHash().Return(importTxHash, nil)
	txs.EXPECT().SetImportTxSig(importTxSig).Return(nil)
	txs.EXPECT().SignedImportTxBytes().Return(importTxHash, nil) // note: here is fake tx bytes, in production it should be generated from a true tx

	txSigner.EXPECT().SignExportTx(ctx, exportTxHash).Return(exportTxSig, nil)
	txSigner.EXPECT().SignImportTx(ctx, importTxHash).Return(importTxSig, nil)

	txIssuer.EXPECT().IssueExportTx(ctx, exportTxHash).Return(ids.ID{}, nil)
	txIssuer.EXPECT().IssueImportTx(ctx, importTxHash).Return(ids.ID{}, nil)

	sigVerifier.EXPECT().VerifyExportTxSig(exportTxHash, exportTxSig).Return(true, nil)
	sigVerifier.EXPECT().VerifyImportTxSig(importTxHash, importTxSig).Return(true, nil)

	tokenPorter := &Porter{
		Txs:         txs,
		TxSigner:    txSigner,
		TxIssuer:    txIssuer,
		SigVerifier: sigVerifier,
	}

	_, err := tokenPorter.SignAndIssueTxs(ctx)
	require.Nil(err)
}

func TestStakeTaskCreatorTestSuite(t *testing.T) {
	suite.Run(t, new(TokenPorterTestSuite))
}
