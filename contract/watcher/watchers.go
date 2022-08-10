package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/contract/watcher"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/network/redialer"
	"github.com/avalido/mpc-controller/utils/network/redialer/adapter"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"time"
)

// Subscribe event: *events.ParticipantAdded
// Subscribe event: *events.KeyGenerated

// Publish event: *events.ParticipantAdded
// Publish event: *events.KeyGenerated
// Publish event: *events.StakeRequestAdded
// Publish event: *events.RequestStarted

type MpcManagerWatchers struct {
	BoundCaller     caller.Caller
	BoundTransactor transactor.Transactor
	ContractAddr    common.Address
	DB              storage.DB
	EthWsURL        string
	KeyGeneratorMPC core.KeygenDoner
	Logger          logger.Logger
	PartiPubKey     storage.PubKey
	Publisher       dispatcher.Publisher

	contractFilterer bind.ContractFilterer
	ethClientCh      chan redialer.Client // todo: handle reconnection
	watcherFactory   *MpcManagerWatcherFactory

	participantAddedWatcher   *watcher.Watcher
	keygenRequestAddedWatcher *watcher.Watcher
	keyGeneratedWatcher       *watcher.Watcher
	stakeRequestAddedWatcher  *watcher.Watcher
	requestStartedWatcher     *watcher.Watcher
}

func (w *MpcManagerWatchers) Init(ctx context.Context) {
	reDialer := adapter.EthClientReDialer{
		Logger:        logger.Default(),
		EthURL:        w.EthWsURL,
		BackOffPolicy: backoff.ExponentialPolicy(0, time.Second, time.Second*10),
	}
	ethClient, ethClientCh, err := reDialer.GetClient(context.Background())
	w.Logger.FatalOnError(err, "Failed to get eth client")
	w.contractFilterer = ethClient.(*ethclient.Client)
	w.ethClientCh = ethClientCh

	boundFilterer, err := contract.BindMpcManagerFilterer(w.ContractAddr, w.contractFilterer)
	w.Logger.FatalOnError(err, "Failed to bind MpcManager filterer")

	watcherFactory := &MpcManagerWatcherFactory{w.Logger, boundFilterer}
	w.watcherFactory = watcherFactory

	err = w.watchParticipantAdded(ctx, nil, [][]byte{w.PartiPubKey})
	w.Logger.FatalOnError(err, "Failed to watch ParticipantAdded")
	err = w.watchRequestStarted(ctx, nil)
	w.Logger.FatalOnError(err, "Failed to watch RequestStarted")
}

func (w *MpcManagerWatchers) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ParticipantAdded:
		groupIDs := [][32]byte{evt.GroupId}
		err := w.watchKeygenRequestAdded(ctx, nil, groupIDs)
		w.Logger.ErrorOnError(err, "Failed to watch KeygenRequestAdded")
		err = w.watchKeyGenerated(ctx, nil, groupIDs)
		w.Logger.ErrorOnError(err, "Failed to watch KeyGenerated")
	case *events.KeyGenerated:
		genPubKeys := [][]byte{evt.PublicKey}
		err := w.watchStakeRequestAdded(ctx, nil, genPubKeys)
		w.Logger.ErrorOnError(err, "Failed to watch StakeRequestAdded")
	}
}

// ParticipantAdded
func (w *MpcManagerWatchers) watchParticipantAdded(ctx context.Context, opts *bind.WatchOpts, pubKeys [][]byte) error {
	participantAddedWatcher, err := w.watcherFactory.NewWatcher(w.processParticipantAdded, opts, EvtParticipantAdded, watcher.QueryFromBytes(pubKeys))
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtParticipantAdded)
	}
	w.participantAddedWatcher = participantAddedWatcher
	err = participantAddedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtParticipantAdded)
}

func (w *MpcManagerWatchers) processParticipantAdded(ctx context.Context, evt interface{}) error { // todo: further process
	// Save participant
	myEvt := evt.(*contract.MpcManagerParticipantAdded)
	participant := storage.Participant{
		PubKey:  myEvt.PublicKey,
		GroupId: myEvt.GroupId,
		Index:   myEvt.Index.Uint64(),
	}
	err := w.DB.SaveModel(ctx, &participant)
	if err != nil {
		return errors.Wrapf(err, "failed to save participant %v", participant)
	}

	// Save group
	rawPubKeys, err := w.BoundCaller.GetGroup(ctx, nil, myEvt.GroupId)
	var pubKeys []storage.PubKey
	for _, rawPubKey := range rawPubKeys {
		pubKeys = append(pubKeys, rawPubKey)
	}
	group := storage.Group{
		ID:    myEvt.GroupId,
		Group: pubKeys,
	}
	err = w.DB.SaveModel(ctx, &group)
	if err != nil {
		return errors.Wrapf(err, "failed to save group %v", group)
	}

	// Publish events.ParticipantAdded
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.ParticipantAdded)(myEvt), nil))
	w.Logger.Debug("Participant added", []logger.Field{{"participant", participant}}...)
	return nil
}

// KeygenRequestAdded
func (w *MpcManagerWatchers) watchKeygenRequestAdded(ctx context.Context, opts *bind.WatchOpts, groupIds [][32]byte) error {
	keygenRequestAddedWatcher, err := w.watcherFactory.NewWatcher(w.processKeygenRequestAdded, opts, EvtKeygenRequestAdded, watcher.QueryFromBytes32(groupIds))
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtKeygenRequestAdded)
	}
	w.keygenRequestAddedWatcher = keygenRequestAddedWatcher
	err = keygenRequestAddedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtKeygenRequestAdded)
}

func (w *MpcManagerWatchers) processKeygenRequestAdded(ctx context.Context, evt interface{}) error {
	myEvt := evt.(*contract.MpcManagerKeygenRequestAdded)

	// Request key generation
	reqId := myEvt.Raw.TxHash.Hex()
	group := storage.Group{ID: myEvt.GroupId}
	err := w.DB.LoadModel(ctx, &group)
	if err != nil {
		return errors.Wrapf(err, "failed to load group %v", group)
	}

	normalized, err := group.Group.CompressPubKeyHexs() // for mpc-server compatibility
	if err != nil {
		return errors.Wrapf(err, "failed to compress participant public keys")
	}

	keyGenReq := &core.KeygenRequest{
		KeygenReqID:            reqId,
		CompressedPartiPubKeys: normalized,
		Threshold:              group.Threshold(),
	}

	res, err := w.KeyGeneratorMPC.KeygenDone(ctx, keyGenReq)
	if err != nil {
		return errors.Wrapf(err, "failed to request key generation %v", keyGenReq)
	}

	// Report generated public key
	genPubKeyHex := res.Result
	dnmGenPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(genPubKeyHex) // for Ethereum compatibility
	if err != nil {
		return errors.Wrapf(err, "failed to decompress generated public key %v", genPubKeyHex)
	}

	participant := storage.Participant{
		PubKey:  hash256.FromBytes(w.PartiPubKey),
		GroupId: myEvt.GroupId,
	}

	err = w.DB.LoadModel(ctx, &participant)
	if err != nil {
		return errors.Wrapf(err, "failed to load participant %v", participant)
	}

	var partiId = participant.ParticipantId()
	_, _, err = w.BoundTransactor.ReportGeneratedKey(ctx, partiId, dnmGenPubKeyBytes)
	if err != nil {
		return errors.Wrapf(err, "failed to report generated public key %v with participant id %v", dnmGenPubKeyBytes, partiId)
	}
	return nil
}

// KeyGenerated
func (w *MpcManagerWatchers) watchKeyGenerated(ctx context.Context, opts *bind.WatchOpts, groupIds [][32]byte) error {
	keyGeneratedWatcher, err := w.watcherFactory.NewWatcher(w.processKeyGenerated, opts, EvtKeyGenerated, watcher.QueryFromBytes32(groupIds))
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtKeyGenerated)
	}
	w.keyGeneratedWatcher = keyGeneratedWatcher
	err = keyGeneratedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtKeyGenerated)
}

func (w *MpcManagerWatchers) processKeyGenerated(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerKeyGenerated)

	// Save generated public key
	genPubKey := storage.GeneratedPublicKey{
		GenPubKey: myEvt.PublicKey,
		GroupId:   myEvt.GroupId,
	}
	err := w.DB.SaveModel(ctx, &genPubKey)
	if err != nil {
		return errors.Wrapf(err, "failed to load generated public key %v", genPubKey)
	}
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.KeyGenerated)(myEvt), nil))
	w.Logger.Debug("Public key generated", []logger.Field{
		{"groupId", genPubKey.GroupId},
		{"genPubKey", genPubKey.GenPubKey},
		{"cChainAddr", genPubKey.GenPubKey.CChainAddress()},
		{"pChainAddr", genPubKey.GenPubKey.PChainAddress()}}...)
	return nil
}

// StakeRequestAdded
func (w *MpcManagerWatchers) watchStakeRequestAdded(ctx context.Context, opts *bind.WatchOpts, pubKeys [][]byte) error {
	stakeRequestAddedWatcher, err := w.watcherFactory.NewWatcher(w.processStakeRequestAdded, opts, EvtStakeRequestAdded, watcher.QueryFromBytes(pubKeys))
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtStakeRequestAdded)
	}
	w.stakeRequestAddedWatcher = stakeRequestAddedWatcher
	err = stakeRequestAddedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtStakeRequestAdded)
}

func (w *MpcManagerWatchers) processStakeRequestAdded(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerStakeRequestAdded)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.StakeRequestAdded)(myEvt), nil))
	return nil
}

// RequestStarted
func (w *MpcManagerWatchers) watchRequestStarted(ctx context.Context, opts *bind.WatchOpts) error {
	requestStartedWatcher, err := w.watcherFactory.NewWatcher(w.processRequestStarted, opts, EvtRequestStarted)
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtRequestStarted)
	}
	w.requestStartedWatcher = requestStartedWatcher
	err = requestStartedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtRequestStarted)
}

func (w *MpcManagerWatchers) processRequestStarted(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerRequestStarted)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.RequestStarted)(myEvt), nil))
	return nil
}
