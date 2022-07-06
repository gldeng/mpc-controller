package rewarding

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/rewarding/export"
	"github.com/avalido/mpc-controller/tasks/rewarding/report"
)

type RewardingMaster struct {
	Logger           logger.Logger
	RewardUTXOGetter chain.RewardUTXOGetter
	Dispatcher       dispatcher.DispatcherClaasic
	MyPubKeyHashHex  string

	// report
	periodEndedChecker    *report.StakingPeriodEndedChecker
	rewardUTXOFetcher     *report.StakingRewardUTXOFetcher
	rewardedStakeReporter *report.RewardedStakeReporter

	// export
	exportRewardReqAddedEvtWatcher  *export.ExportRewardRequestAddedEventWatcher
	exportRewardJoiner              *export.ExportRewardRequestJoiner
	exportRewarReqStartedEvtWatcher *export.ExportRewardRequestStartedEventWatcher
	rewardExporter                  *export.StakingRewardExporter
}

func (s *RewardingMaster) Start(ctx context.Context) error {
	s.subscribe()
	<-ctx.Done()
	return nil
}

func (s *RewardingMaster) subscribe() {
	periodEndedChecker := report.StakingPeriodEndedChecker{
		Publisher: s.Dispatcher,
	}

	rewardUTXOFetcher := report.StakingRewardUTXOFetcher{
		Logger:           s.Logger,
		Publisher:        s.Dispatcher,
		RewardUTXOGetter: s.RewardUTXOGetter,
	}
	//
	//rewardedStakeReporter := report.RewardedStakeReporter{
	//	Logger:           s.Logger,
	//	MyPubKeyHashHex string
	//
	//	Publisher:        s.Dispatcher,
	//
	//	Cache cache.ICache
	//
	//	Signer *bind.TransactOpts
	//
	//	ContractAddr common.Address
	//	Transactor   bind.ContractTransactor
	//
	//	Receipter chain.Receipter
	//}
	//
	s.periodEndedChecker = &periodEndedChecker
	s.rewardUTXOFetcher = &rewardUTXOFetcher
	//s.rewardUTXOFetcher =
	//
	//	s.exportRewardReqAddedEvtWatcher =
	//		s.exportRewardJoiner =
	//			s.exportRewarReqStartedEvtWatcher =
	//				s.rewardExporter

	s.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, s.periodEndedChecker) // Emit event: *events.StakingPeriodEndedEvent

	s.Dispatcher.Subscribe(&events.StakingPeriodEndedEvent{}, s.rewardUTXOFetcher) // Emit event: *events.RewardUTXOsFetchedEvent
}
