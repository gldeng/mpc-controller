package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/contract/watcher"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/network/redialer"
	"github.com/avalido/mpc-controller/utils/network/redialer/adapter"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"time"
)

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

	groupIDs   map[common.Hash]struct{}
	genPubKeys map[string]storage.PubKey
}

func (w *MpcManagerWatchers) Init(ctx context.Context) {
	w.groupIDs = make(map[common.Hash]struct{})
	w.genPubKeys = make(map[string]storage.PubKey)
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

	// read stored group and watch for keygenRequestAdded and keyGenerated contract events.
	groupBytesArr, err := w.DB.List(ctx, storage.KeyPrefixGroup)
	w.Logger.FatalOnTrue(err != nil && !errors.Is(err, badger.ErrKeyNotFound), "Failed to query group")
	w.Logger.InfoOnTrue(len(groupBytesArr) == 0, fmt.Sprintf("No group loaded from db by key prefix %v", string(storage.KeyPrefixGroup)))
	for _, groupBytes := range groupBytesArr {
		var group storage.Group
		err := json.Unmarshal(groupBytes, &group)
		w.Logger.FatalOnError(err, "Failed to unmarshal group")
		w.Logger.Info("Read group from storage", []logger.Field{{"groupId", group.ID.String()}}...)
		groupIDs := [][32]byte{group.ID}
		err = w.watchKeygenRequestAdded(ctx, nil, groupIDs)
		w.Logger.FatalOnError(err, "Failed to watch KeygenRequestAdded")
		err = w.watchKeyGenerated(ctx, nil, groupIDs)
		w.Logger.FatalOnError(err, "Failed to watch KeyGenerated")
		w.groupIDs[group.ID] = struct{}{}
	}

	// read stored generated key and watch for stakeRequestAdded contract event.
	genKeyBytesArr, err := w.DB.List(ctx, storage.KeyPrefixGeneratedPublicKey)
	w.Logger.FatalOnTrue(err != nil && !errors.Is(err, badger.ErrKeyNotFound), "Failed to query generated key")
	w.Logger.InfoOnTrue(len(groupBytesArr) == 0, fmt.Sprintf("No generated public key loaded from db by key prefix %v", string(storage.KeyPrefixGeneratedPublicKey)))
	for _, genKeyBytes := range genKeyBytesArr {
		var genPubKey storage.GeneratedPublicKey
		err := json.Unmarshal(genKeyBytes, &genPubKey)
		w.Logger.FatalOnError(err, "Failed to unmarshal generated key")
		genKey := [][]byte{genPubKey.GenPubKey}
		err = w.watchStakeRequestAdded(ctx, nil, genKey)
		w.Logger.FatalOnError(err, "Failed to watch StakeRequestAdded")

		cChainAddr, err := genPubKey.GenPubKey.CChainAddress()
		w.Logger.FatalOnError(err, fmt.Sprintf("Failed to get C-Chain address from %v", genPubKey.GenPubKey))
		pChainAddr, err := genPubKey.GenPubKey.PChainAddress()
		w.Logger.FatalOnError(err, fmt.Sprintf("Failed to get P-Chain address from %v", genPubKey.GenPubKey))
		w.Logger.Info("Public key loaded", []logger.Field{
			{"groupId", genPubKey.GroupId},
			{"genPubKey", genPubKey.GenPubKey},
			{"cChainAddr", cChainAddr},
			{"pChainAddr", pChainAddr}}...)
		w.genPubKeys[genPubKey.GenPubKey.String()] = genPubKey.GenPubKey
	}

	// rewatch contract event upon ws reconnection
	go w.reWatch(ctx)
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
	w.Logger.Info("Watching contract event", []logger.Field{{"event", EvtParticipantAdded}}...)
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

	w.groupIDs[group.ID] = struct{}{}

	// Publish events.ParticipantAdded
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.ParticipantAdded)(myEvt), nil))
	w.Logger.Info("Participant added", []logger.Field{{"participant", participant}}...)
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
	w.Logger.Info("Watching contract event", []logger.Field{{"event", EvtKeygenRequestAdded}}...)
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
		ReqID:                  string(events.ReqIDPrefixKeygen) + reqId,
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
	w.Logger.Info("Reported generated public key", []logger.Field{{"genPubKey", bytes.BytesToHex(dnmGenPubKeyBytes)}}...)
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
	w.Logger.Info("Watching contract event", []logger.Field{{"event", EvtKeyGenerated}}...)
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
	w.genPubKeys[genPubKey.GenPubKey.String()] = genPubKey.GenPubKey
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.KeyGenerated)(myEvt), nil))
	cChainAddr, err := genPubKey.GenPubKey.CChainAddress()
	if err != nil {
		return errors.Wrapf(err, "failed to get C-Chain address from %v", genPubKey.GenPubKey)
	}
	pChainAddr, err := genPubKey.GenPubKey.PChainAddress()
	if err != nil {
		return errors.Wrapf(err, "failed to get P-Chain address from %v", genPubKey.GenPubKey)
	}
	w.Logger.Info("Public key generated", []logger.Field{
		{"groupId", genPubKey.GroupId},
		{"genPubKey", genPubKey.GenPubKey},
		{"cChainAddr", cChainAddr},
		{"pChainAddr", pChainAddr}}...)
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
	w.Logger.Info("Watching contract event", []logger.Field{{"event", EvtStakeRequestAdded}}...)
	return errors.Wrapf(err, "failed to watch %v", EvtStakeRequestAdded)
}

func (w *MpcManagerWatchers) processStakeRequestAdded(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerStakeRequestAdded)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.StakeRequestAdded)(myEvt), nil))
	w.Logger.Info("Stake request added", []logger.Field{{"stakeReqAdded", myEvt}}...)
	prom.StakeRequestAdded.Inc()
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
	w.Logger.Info("Watching contract event", []logger.Field{{"event", EvtRequestStarted}}...)
	return errors.Wrapf(err, "failed to watch %v", EvtRequestStarted)
}

func (w *MpcManagerWatchers) processRequestStarted(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerRequestStarted)
	reqHash := (storage.RequestHash)(myEvt.RequestHash)
	indices := (*storage.Indices)(myEvt.ParticipantIndices)
	w.Logger.Info("Request started", []logger.Field{{"reqStarted",
		fmt.Sprintf("reqHash:%v, partiIndices:%v", reqHash.String(), indices.Indices())}}...)
	switch {
	case reqHash.IsTaskType(storage.TaskTypStake):
		stakeReq := storage.StakeRequest{}
		joinReq := storage.JoinRequest{
			ReqHash: reqHash,
			Args:    &stakeReq,
		}
		if err := w.DB.LoadModel(ctx, &joinReq); err != nil {
			break
		}
		if !joinReq.PartiId.Joined(myEvt.ParticipantIndices) {
			break
		}
		cmpGenPubKeyHex, joinedCmpPartiPubKeys, err := w.getKeyInfo(ctx, stakeReq.GeneratedPublicKey, indices)
		if err != nil {
			w.Logger.ErrorOnError(err, "Failed to get key info")
			break
		}
		evt := events.RequestStarted{
			ReqHash:                &reqHash,
			TaskType:               storage.TaskTypStake,
			PartiIndices:           indices,
			JoinedReq:              &joinReq,
			CompressedPartiPubKeys: joinedCmpPartiPubKeys,
			CompressedGenPubKeyHex: cmpGenPubKeyHex,
			Raw:                    myEvt.Raw,
		}
		w.Publisher.Publish(ctx, dispatcher.NewEvtObj(&evt, nil))
		w.Logger.Info("Stake request started", []logger.Field{{"stakeReqStarted", evt}}...)
		prom.StakeRequestStarted.Inc()
	case reqHash.IsTaskType(storage.TaskTypReturn):
		utxoExportReq := storage.ExportUTXORequest{}
		joinReq := storage.JoinRequest{
			ReqHash: reqHash,
			Args:    &utxoExportReq,
		}

		if err := w.DB.LoadModel(ctx, &joinReq); err != nil {
			//w.Logger.DebugOnError(err, "No JoinRequest load for UTXO export", []logger.Field{{"reqHash", joinReq.ReqHash}}...)
			break
		}
		if !joinReq.PartiId.Joined(myEvt.ParticipantIndices) {
			//w.Logger.Debug("Not joined UTXO export request", []logger.Field{{"reqHash", myEvt.ReqHash}}...)
			break
		}
		w.Publisher.Publish(ctx, dispatcher.NewEvtObj(&events.RequestStarted{indices, &joinReq, myEvt.Raw}, nil))
		w.Logger.Info("Return request started", []logger.Field{{"returnReqStarted",
			fmt.Sprintf("reqHash:%v, partiIndices:%v, returnReq:%+v", reqHash.String(), indices.Indices(), utxoExportReq)}}...)
		prom.UTXOExportRequestStarted.Inc()
	}
	return nil
}

func (w *MpcManagerWatchers) getKeyInfo(ctx context.Context, genPubKey *storage.GeneratedPublicKey, partiIndices *storage.Indices) (string, []string, error) {
	// Get generated key and participant keys
	cmpGenPubKeyHex, err := genPubKey.GenPubKey.CompressPubKeyHex()
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to compress public key")
	}

	group := storage.Group{
		ID: genPubKey.GroupId,
	}
	if err := w.DB.LoadModel(ctx, &group); err != nil {
		return "", nil, errors.Wrapf(err, "failed to load group")
	}

	cmpPartiPubKeys, err := group.Group.CompressPubKeyHexs()
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to comporess participant public keys")
	}

	var joinedCmpPartiPubKeys []string
	indices := partiIndices.Indices()
	for _, index := range indices {
		joinedCmpPartiPubKeys = append(joinedCmpPartiPubKeys, cmpPartiPubKeys[index-1])
	}

	return cmpGenPubKeyHex, joinedCmpPartiPubKeys, nil
}

func (w *MpcManagerWatchers) reWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ethClient := <-w.ethClientCh:
			w.Logger.Info("EthClient reconnected")
			w.closeSubWatchers()
			w.contractFilterer = ethClient.(*ethclient.Client)
			boundFilterer, err := contract.BindMpcManagerFilterer(w.ContractAddr, w.contractFilterer)
			w.Logger.ErrorOnError(err, "Failed to rebind MpcManager filterer")
			if err != nil {
				break
			}
			watcherFactory := &MpcManagerWatcherFactory{w.Logger, boundFilterer}
			w.watcherFactory = watcherFactory

			err = w.watchParticipantAdded(ctx, nil, [][]byte{w.PartiPubKey})
			w.Logger.ErrorOnError(err, "Failed to rewatch ParticipantAdded")
			err = w.watchRequestStarted(ctx, nil)
			w.Logger.ErrorOnError(err, "Failed to rewatch RequestStarted")

			for groupID, _ := range w.groupIDs {
				groupIDs := [][32]byte{groupID}
				err = w.watchKeygenRequestAdded(ctx, nil, groupIDs)
				w.Logger.ErrorOnError(err, "Failed to rewatch KeygenRequestAdded")
				err = w.watchKeyGenerated(ctx, nil, groupIDs)
				w.Logger.ErrorOnError(err, "Failed to rewatch KeyGenerated")
			}
			if err != nil {
				break
			}

			for _, genPubKey := range w.genPubKeys {
				genKey := [][]byte{genPubKey}
				err = w.watchStakeRequestAdded(ctx, nil, genKey)
				w.Logger.ErrorOnError(err, "Failed to rewatch StakeRequestAdded")
			}
			if err != nil {
				break
			}
			w.Logger.Info("Contract events rewatched")
		}
	}
}

func (w *MpcManagerWatchers) closeSubWatchers() {
	w.participantAddedWatcher.Close()
	w.keygenRequestAddedWatcher.Close()
	w.keyGeneratedWatcher.Close()
	w.stakeRequestAddedWatcher.Close()
	w.requestStartedWatcher.Close()
}
