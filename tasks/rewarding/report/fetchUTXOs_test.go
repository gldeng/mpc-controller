package report

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	chainMocks "github.com/avalido/mpc-controller/chain/mocks"
	"github.com/avalido/mpc-controller/dispatcher"
	dispatcherMocks "github.com/avalido/mpc-controller/dispatcher/mocks"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type StakingRewardUTXOFetcherTestSuite struct {
	suite.Suite
}

func (suite *StakingRewardUTXOFetcherTestSuite) SetupTest() {}

func (suite *StakingRewardUTXOFetcherTestSuite) TestSignTx() {
	require := suite.Require()
	myDispatcher := dispatcherMocks.NewDispatcherrer(suite.T())
	myRewardUTXOGetter := chainMocks.NewRewardUTXOGetter(suite.T())

	periodEndedEvt := &events.StakingPeriodEndedEvent{
		AddDelegatorTxID: ids.ID{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	periodEndedEvtObj := dispatcher.NewRootEventObject("StakingPeriodEndedChecker", periodEndedEvt, ctx)

	rewardUTXOsFetchedEvt := &events.RewardUTXOsFetchedEvent{}

	myRewardUTXOGetter.EXPECT().GetRewardUTXOs(ctx, mock.AnythingOfType("*api.GetTxArgs")).Return([][]byte{[]byte("hello")}, nil) //todo: generate fake UTXOs

	myDispatcher.EXPECT().Publish(ctx, mock.AnythingOfType("*dispatcher.EventObject")).Run(func(ctx context.Context, evtObj *dispatcher.EventObject) {
		rewardUTXOsFetchedEvt = evtObj.Event.(*events.RewardUTXOsFetchedEvent)
	})

	checker := &StakingRewardUTXOFetcher{
		Logger:           logger.Default(),
		Publisher:        myDispatcher,
		RewardUTXOGetter: myRewardUTXOGetter,
	}

	checker.Do(ctx, periodEndedEvtObj)

	<-ctx.Done()

	require.Equal(periodEndedEvt.AddDelegatorTxID, rewardUTXOsFetchedEvt.AddDelegatorTxID)
	require.True(len(rewardUTXOsFetchedEvt.RewardUTXOs) > 0)

	myDispatcher.AssertNumberOfCalls(suite.T(), "Publish", 1)
	myRewardUTXOGetter.AssertNumberOfCalls(suite.T(), "GetRewardUTXOs", 1)
}

func TestStakeTaskCreatorTestSuite(t *testing.T) {
	suite.Run(t, new(StakingRewardUTXOFetcherTestSuite))
}
