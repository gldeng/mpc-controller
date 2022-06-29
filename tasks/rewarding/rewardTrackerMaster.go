package rewarding

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/rewarding/stakingRewardUTXOFetcher"
)

type RewardingMaster struct {
	Logger           logger.Logger
	RewardUTXOGetter chain.RewardUTXOGetter
	Dispatcher       dispatcher.DispatcherClaasic

	rewardUTXOTracker *stakingRewardUTXOFetcher.StakingRewardUTXOFetcher
}

func (s *RewardingMaster) Start(ctx context.Context) error {
	s.subscribe()
	<-ctx.Done()
	return nil
}

func (s *RewardingMaster) subscribe() {
	rewardUTXOTracker := stakingRewardUTXOFetcher.StakingRewardUTXOFetcher{
		Logger:           s.Logger,
		RewardUTXOGetter: s.RewardUTXOGetter,
	}

	s.rewardUTXOTracker = &rewardUTXOTracker

	s.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, s.rewardUTXOTracker)
}
