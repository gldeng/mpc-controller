package rewarding

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/rewarding/track/checkPeriod"
	"github.com/avalido/mpc-controller/tasks/rewarding/track/fetchUTXOs"
)

type RewardingMaster struct {
	Logger           logger.Logger
	RewardUTXOGetter chain.RewardUTXOGetter
	Dispatcher       dispatcher.DispatcherClaasic

	periodEndedChecker *checkPeriod.StakingPeriodEndedChecker
	rewardUTXOFetcher  *fetchUTXOs.StakingRewardUTXOFetcher
}

func (s *RewardingMaster) Start(ctx context.Context) error {
	s.subscribe()
	<-ctx.Done()
	return nil
}

func (s *RewardingMaster) subscribe() {
	periodEndedChecker := checkPeriod.StakingPeriodEndedChecker{
		Publisher: s.Dispatcher,
	}

	rewardUTXOFetcher := fetchUTXOs.StakingRewardUTXOFetcher{
		Logger:           s.Logger,
		Publisher:        s.Dispatcher,
		RewardUTXOGetter: s.RewardUTXOGetter,
	}

	s.periodEndedChecker = &periodEndedChecker
	s.rewardUTXOFetcher = &rewardUTXOFetcher

	s.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, s.periodEndedChecker) // Emit event: *events.StakingPeriodEndedEvent

	s.Dispatcher.Subscribe(&events.StakingPeriodEndedEvent{}, s.rewardUTXOFetcher) // Emit event: *events.RewardUTXOsFetchedEvent
}
