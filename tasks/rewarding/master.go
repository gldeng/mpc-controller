package rewarding

import (
	"context"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/rewarding/porter"
	"github.com/avalido/mpc-controller/tasks/rewarding/tracker"
	"github.com/avalido/mpc-controller/tasks/rewarding/watcher"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Master struct {
	CChainIssueClient chain.CChainIssuer
	Cache             cache.ICache
	ContractAddr      common.Address
	Dispatcher        dispatcher.Dispatcher
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
	utxoFetchedEventCache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	utxoExportedEventCache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	utxoTracker := tracker.UTXOTracker{
		UTXOsFetchedEventCache: utxoFetchedEventCache,
		UTXOExportedEventCache: utxoExportedEventCache,
		ContractAddr:           m.ContractAddr,
		Logger:                 m.Logger,
		PChainClient:           m.PChainClient,
		Publisher:              m.Dispatcher,
		Receipter:              m.Receipter,
		Signer:                 m.Signer,
		Transactor:             m.Transactor,
	}

	exportUTXOReqEvtWatcher := watcher.ExportUTXORequestWatcher{
		ContractAddr: m.ContractAddr,
		Logger:       m.Logger,
		Publisher:    m.Dispatcher,
	}

	utxoPorter := porter.UTXOPorter{
		UTXOsFetchedEventCache: utxoFetchedEventCache,
		UTXOExportedEventCache: utxoExportedEventCache,
		CChainIssueClient:      m.CChainIssueClient,
		Cache:                  m.Cache,
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

	m.Dispatcher.Subscribe(&events.ReportedGenPubKeyEvent{}, m.utxoTracker)

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.exportUTXOReqEvtWatcher)
	m.Dispatcher.Subscribe(&events.ReportedGenPubKeyEvent{}, m.exportUTXOReqEvtWatcher)

	m.Dispatcher.Subscribe(&events.ReportedGenPubKeyEvent{}, m.utxoPorter)
	m.Dispatcher.Subscribe(&events.ExportUTXORequestEvent{}, m.utxoPorter)
}
