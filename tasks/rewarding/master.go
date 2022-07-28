package rewarding

import (
	"context"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/rewarding/porter"
	"github.com/avalido/mpc-controller/tasks/rewarding/tracker"
	"github.com/avalido/mpc-controller/tasks/rewarding/watcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

type Master struct {
	CChainIssueClient chain.CChainIssuer
	Cache             cache.ICache
	ContractAddr      common.Address
	Dispatcher        dispatcher.DispatcherClaasic
	Logger            logger.Logger
	MyPubKeyHashHex   string
	PChainClient      platformvm.Client
	Receipter         chain.Receipter
	RewardUTXOGetter  chain.RewardUTXOGetter
	SignDoner         core.SignDoner
	Signer            *bind.TransactOpts
	Transactor        bind.ContractTransactor
	chain.NetworkContext

	// track
	utxoTracker *tracker.UTXOTracker

	// export
	exportUTXOReqEvtWatcher *watcher.ExportUTXORequestWatcher
	utxoPorter              *porter.UTXOPorter
}

func (m *Master) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *Master) subscribe() {
	utxosFetchedEvent := new(sync.Map)

	utxoTracker := tracker.UTXOTracker{
		ContractAddr:           m.ContractAddr,
		Logger:                 m.Logger,
		PChainClient:           m.PChainClient,
		Publisher:              m.Dispatcher,
		Receipter:              m.Receipter,
		Signer:                 m.Signer,
		Transactor:             m.Transactor,
		UTXOsFetchedEventCache: utxosFetchedEvent,
	}

	exportUTXOReqEvtWatcher := watcher.ExportUTXORequestWatcher{
		ContractAddr: m.ContractAddr,
		Logger:       m.Logger,
		Publisher:    m.Dispatcher,
	}

	utxoPorter := porter.UTXOPorter{
		CChainIssueClient:      m.CChainIssueClient,
		Cache:                  m.Cache,
		UTXOsFetchedEventCache: utxosFetchedEvent,
		Logger:                 m.Logger,
		MyPubKeyHashHex:        m.MyPubKeyHashHex,
		NetworkContext:         m.NetworkContext,
		PChainIssueClient:      m.PChainClient,
		Publisher:              m.Dispatcher,
		SignDoner:              m.SignDoner,
	}

	m.utxoTracker = &utxoTracker
	m.exportUTXOReqEvtWatcher = &exportUTXOReqEvtWatcher
	m.utxoPorter = &utxoPorter

	m.Dispatcher.Subscribe(&events.ReportedGenPubKeyEvent{}, m.utxoTracker) // Emit event: *events.UTXOsFetchedEventCache; *events.UTXOReportedEvent
	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.exportUTXOReqEvtWatcher)
	m.Dispatcher.Subscribe(&events.ReportedGenPubKeyEvent{}, m.exportUTXOReqEvtWatcher) // Emit event: *contract.ExportUTXORequestEvent

	m.Dispatcher.Subscribe(&events.ExportUTXORequestEvent{}, m.utxoPorter) // Emit event: *events.UTXOExportedEvent
}
