package report

//
//import (
//	"context"
//	"github.com/ava-labs/avalanchego/ids"
//	chainMocks "github.com/avalido/mpc-controller/chain/mocks"
//	contractMocks "github.com/avalido/mpc-controller/contract/mocks"
//	"github.com/avalido/mpc-controller/dispatcher"
//	dispatcherMocks "github.com/avalido/mpc-controller/dispatcher/mocks"
//	"github.com/avalido/mpc-controller/events"
//	"github.com/avalido/mpc-controller/logger"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/suite"
//	"testing"
//	"time"
//)
//
//type StakingRewardUTXOReporterTestSuite struct {
//	suite.Suite
//}
//
//func (suite *StakingRewardUTXOReporterTestSuite) SetupTest() {}
//
//func (suite *StakingRewardUTXOReporterTestSuite) TestSignTx() {
//	require := suite.Require()
//
//	myTransactor := contractMocks.NewTransactorReportRewardUTXOs(suite.T())
//	myReceipter := chainMocks.NewReceipter(suite.T())
//	myDispatcher := dispatcherMocks.NewDispatcherrer(suite.T())
//
//	utxoFetchedEvt := &events.RewardUTXOsFetchedEvent{ // todo: mock data
//		TxID: ids.ID{},
//		NativeUTXOs:      nil,
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//	defer cancel()
//
//	myTransactor.EXPECT().ReportRewardUTXOs(ctx, [32]byte(utxoFetchedEvt.TxID), mock.AnythingOfType("[]string")).Return(&types.Transaction{}, nil) // todo: mock data
//
//	myReceipter.EXPECT().TransactionReceipt(ctx, mock.AnythingOfType("common.Hash")).Return(&types.Receipt{Status: 1}, nil)
//
//	utxoFetchEvtObj := dispatcher.NewRootEventObject("StakingRewardUTXOFetcher", utxoFetchedEvt, ctx)
//
//	rewardUTXOsReportedEvt := &events.RewardedStakeReportedEvent{}
//
//	myDispatcher.EXPECT().Publish(ctx, mock.AnythingOfType("*dispatcher.EventObject")).Run(func(ctx context.Context, evtObj *dispatcher.EventObject) {
//		rewardUTXOsReportedEvt = evtObj.Event.(*events.RewardedStakeReportedEvent)
//	})
//
//	reporter := &UTXOReporter{
//		Logger:     logger.Default(),
//		Publisher:  myDispatcher,
//		Transactor: myTransactor,
//		Receipter:  myReceipter,
//	}
//
//	reporter.Do(ctx, utxoFetchEvtObj)
//
//	<-ctx.Done()
//
//	require.Equal(utxoFetchedEvt.TxID, rewardUTXOsReportedEvt.TxID)
//
//	myTransactor.AssertNumberOfCalls(suite.T(), " ReportRewardUTXOs", 1)
//	myReceipter.AssertNumberOfCalls(suite.T(), " TransactionReceipt", 1)
//	myDispatcher.AssertNumberOfCalls(suite.T(), "Publish", 1)
//}
//
//func TestStakingRewardUTXOReporterTestSuite(t *testing.T) {
//	suite.Run(t, new(StakingRewardUTXOReporterTestSuite))
//}
