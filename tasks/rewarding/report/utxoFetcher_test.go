package report

//
//import (
//	"context"
//	"github.com/avalido/mpc-controller/dispatcher"
//	dispatcherMocks "github.com/avalido/mpc-controller/dispatcher/mocks"
//	"github.com/avalido/mpc-controller/events"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/suite"
//	"testing"
//	"time"
//)
//
//type StakingPeriodEndedCheckerTestSuite struct {
//	suite.Suite
//}
//
//func (suite *StakingPeriodEndedCheckerTestSuite) SetupTest() {}
//
//func (suite *StakingPeriodEndedCheckerTestSuite) TestSignTx() {
//	require := suite.Require()
//	myDispatcher := dispatcherMocks.NewDispatcherrer(suite.T())
//
//	endTime := time.Now().Add(time.Second * 5)
//
//	stakingDoneEvt := &events.StakingTaskDoneEvent{
//		EndTime: uint64(endTime.Unix()),
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), endTime.Add(time.Second*emitAfterSeconds).Sub(time.Now()))
//	defer cancel()
//
//	stakingDoneEvtObj := dispatcher.NewRootEventObject("UTXOFetcher", stakingDoneEvt, ctx)
//
//	StakingEndedEvt := &events.StakingPeriodEndedEvent{}
//
//	myDispatcher.EXPECT().Publish(ctx, mock.AnythingOfType("*dispatcher.EventObject")).Run(func(ctx context.Context, evtObj *dispatcher.EventObject) {
//		StakingEndedEvt = evtObj.Event.(*events.StakingPeriodEndedEvent)
//	})
//
//	checker := &UTXOFetcher{
//		Publisher: myDispatcher,
//	}
//
//	checker.Do(ctx, stakingDoneEvtObj)
//
//	<-ctx.Done()
//
//	require.Equal(stakingDoneEvt.EndTime, StakingEndedEvt.EndTime)
//
//	myDispatcher.AssertNumberOfCalls(suite.T(), "Publish", 1)
//}
//
//func TestStakingPeriodEndedCheckerTestSuite(t *testing.T) {
//	suite.Run(t, new(StakingPeriodEndedCheckerTestSuite))
//}
