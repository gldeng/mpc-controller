package rewarding

import (
	"context"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/tasks/rewarding/porter"
	"github.com/avalido/mpc-controller/tasks/rewarding/tracker"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/dgraph-io/ristretto"
)

type Master struct {
	BoundCaller            caller.Caller
	BoundTransactor        transactor.Transactor
	ClientPChain           platformvm.Client
	DB                     storage.DB
	Dispatcher             dispatcher.Dispatcher
	IssuerCChain           chain.CChainIssuer
	IssuerPChain           chain.PChainIssuer
	Logger                 logger.Logger
	NetWorkCtx             chain.NetworkContext
	PartiPubKey            storage.PubKey
	SignerMPC              core.SignDoner
	UTXOExportedEventCache *ristretto.Cache
	UTXOsFetchedEventCache *ristretto.Cache

	// track
	utxoTracker *tracker.UTXOTracker

	// export
	utxoPorter *porter.UTXOPorter
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
		BoundTransactor:        m.BoundTransactor,
		ClientPChain:           m.ClientPChain,
		DB:                     m.DB,
		Logger:                 m.Logger,
		PartiPubKey:            m.PartiPubKey,
		UTXOExportedEventCache: utxoExportedEventCache,
		UTXOsFetchedEventCache: utxoFetchedEventCache,
	}

	utxoPorter := porter.UTXOPorter{
		BoundCaller:            m.BoundCaller,
		DB:                     m.DB,
		IssuerCChain:           m.IssuerCChain,
		IssuerPChain:           m.IssuerPChain,
		Logger:                 m.Logger,
		NetWorkCtx:             m.NetWorkCtx,
		SignerMPC:              m.SignerMPC,
		UTXOExportedEventCache: utxoExportedEventCache,
		UTXOsFetchedEventCache: utxoFetchedEventCache,
	}

	m.utxoTracker = &utxoTracker
	m.utxoPorter = &utxoPorter

	m.Dispatcher.Subscribe(&events.KeyGenerated{}, m.utxoTracker)
	m.Dispatcher.Subscribe(&events.RequestStarted{}, m.utxoPorter)
}
