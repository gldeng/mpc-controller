package tracker

import (
	"context"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
)

type RewardTrackerMaster struct {
	Logger           logger.Logger
	RewardUTXOGetter RewardUTXOGetter
	Dispatcher       dispatcher.DispatcherClaasic

	rewardUTXOTracker *RewardUTXOTracker
}

func (s *RewardTrackerMaster) Start(ctx context.Context) error {
	s.subscribe()
	<-ctx.Done()
	return nil
}

func (s *RewardTrackerMaster) subscribe() {
	rewardUTXOTracker := RewardUTXOTracker{
		Logger:           s.Logger,
		RewardUTXOGetter: s.RewardUTXOGetter,
	}

	s.rewardUTXOTracker = &rewardUTXOTracker

	s.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, s.rewardUTXOTracker)
}
