package keygen

import (
	"context"
	ctlPk "github.com/avalido/mpc-controller"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
// Keygen implementation

var _ ctlPk.MpcControllerService = (*Keygen)(nil)

type ReportKeyTx struct {
	groupId            [32]byte
	myIndex            *big.Int
	generatedPublicKey []byte
}

// Keygen instance watches ParticipantAdded event emitted from MpcManager contract,
// which will result in local persistence of corresponding ParticipantInfo and GroupInfo datum.
// todo: consider break its responsibility
type Keygen struct {
	logger.Logger
	PubKeyHashHex string

	ctlPk.MpcClientKeygen
	ctlPk.MpcClientResult

	ctlPk.TransactorReportGeneratedKey

	ctlPk.WatcherKeygenRequestAdded // MpcManager filter

	ctlPk.StorerGetGroupIds
	ctlPk.StorerLoadParticipantInfo

	ctlPk.StorerLoadKeygenRequestInfo
	ctlPk.StorerStoreGeneratedPubKeyInfo

	ctlPk.StorerLoadGroupInfo
	ctlPk.StorerStoreKeygenRequestInfo

	ctlPk.EthClientTransactionReceipt

	keygenRequestAddedEvt chan *contract.MpcManagerKeygenRequestAdded
	pendingKeygenRequests map[string]*core.KeygenRequest
	pendingReports        map[common.Hash]*ReportKeyTx

	signer *bind.TransactOpts
}

// todo: deal with ticking for checking keygen result, use pipeline pattern instead

func (k *Keygen) Start(ctx context.Context) error {
	// Watch KeygenRequestAdded event
	go func() {
		err := k.watchKeygenRequestAdded(ctx)
		k.ErrorOnError(err, "Got an error to watch KeygenRequestAdded event")
	}()

	// Action upon KeygenRequestAdded
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-k.keygenRequestAddedEvt:
			err := k.onKeygenRequestAdded(ctx, evt)
			k.ErrorOnError(err, "Failed to process ParticipantAdded event")
		}
	}
}

func (k *Keygen) watchKeygenRequestAdded(ctx context.Context) error {
	// Subscribe KeygenRequestAdded event
	groupIds, err := k.GetGroupIds(k.PubKeyHashHex)
	if err != nil {
		return errors.WithStack(err)
	}

	sink, err := k.WatchKeygenRequestAdded(groupIds)
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch KeygenRequestAdded event
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-sink:
			k.WarnOnNotOk(ok, "Retrieve nothing from event channel of KeygenRequestAdded")
			if ok {
				k.keygenRequestAddedEvt <- evt
			}
		}
	}
}

func (k *Keygen) onKeygenRequestAdded(ctx context.Context, evt *contract.MpcManagerKeygenRequestAdded) error {
	groupIdHex := common.Bytes2Hex(evt.GroupId[:])

	groupInfo, err := k.LoadGroupInfo(groupIdHex)
	if err != nil {
		return errors.WithStack(err)
	}

	reqIdHex := evt.Raw.TxHash.Hex()
	partPubKeyHexs := groupInfo.PartPubKeyHexs
	request := &core.KeygenRequest{
		RequestId:       reqIdHex,
		ParticipantKeys: partPubKeyHexs,

		Threshold: groupInfo.Threshold,
	}

	err = k.Keygen(ctx, request)
	if err != nil {
		return errors.Wrapf(err, "failed to send keygen request to mpc-server")
	}
	k.pendingKeygenRequests[request.RequestId] = request

	keygenReqInfo := storage.KeygenRequestInfo{
		RequestIdHex:   reqIdHex,
		GroupIdHex:     groupIdHex,
		RequestAddedAt: time.Now(),
	}

	err = k.StoreKeygenRequestInfo(&keygenReqInfo)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// todo: refactor ti into piple pattern.
func (k *Keygen) checkPendingReports(ctx context.Context) error {
	var done []common.Hash
	var retry []common.Hash
	for txHash, _ := range k.pendingReports {
		rcp, err := k.TransactionReceipt(ctx, txHash)
		if err == nil {
			if rcp.Status == 1 {
				done = append(done, txHash)
			} else {
				retry = append(retry, txHash)
			}
		}
	}
	// TODO: Figure out why tx fails
	// Suspect due to contention between different users, for now make retry random
	sampledRetry := sample(retry)

	for _, txHash := range sampledRetry {
		req := k.pendingReports[txHash]
		groupId, ind, pubkey := req.groupId, req.myIndex, req.generatedPublicKey
		tx, err := k.ReportGeneratedKey(k.signer, groupId, ind, pubkey)
		k.pendingReports[tx.Hash()] = &ReportKeyTx{
			groupId:            groupId,
			myIndex:            ind,
			generatedPublicKey: pubkey,
		}

		if err != nil {
			k.Error("Failed to report generated key", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		k.Info("Retry reporting key.", []logger.Field{
			{"groupId", common.Bytes2Hex(groupId[:])},
			{"myIndex", ind},
			{"pubKey", common.Bytes2Hex(pubkey)},
			{"txHash", tx.Hash()}}...)

	}
	for _, txHash := range sampledRetry {
		delete(k.pendingReports, txHash)
	}
	for _, txHash := range done {
		delete(k.pendingReports, txHash)
	}
	return nil
}

func (k *Keygen) checkKeygenResult(ctx context.Context, requestId string) error {
	// Query result from mpc-server
	result, err := k.Result(ctx, requestId) // todo: add shared context to task manager
	if err != nil {
		return errors.WithStack(err)
	}

	if result.RequestStatus != "DONE" {
		k.Debug("Key hasn't been generated yet",
			[]logger.Field{{"reqId", requestId}, {"reqStatus", result.RequestStatus}}...)
		return nil
	}

	// Deal with crypto values
	dnmGenPubKeyBytes, err := myCrypto.DenormalizePubKeyFromHex(result.Result) // for Ethereum compatibility
	if err != nil {
		return errors.WithStack(err)
	}
	pubKeyHash := crypto.Keccak256Hash(dnmGenPubKeyBytes) // digest to identify generated public key

	// Load pre-stored corresponding participant and keygen request information
	keyGenInfo, err := k.LoadKeygenRequestInfo(requestId) // keygen request info
	if err != nil {
		return errors.WithStack(err)
	}
	partyInfo, err := k.LoadParticipantInfo(k.PubKeyHashHex, keyGenInfo.GroupIdHex) // participant info
	if err != nil {
		return errors.WithStack(err)
	}

	groupIdRaw := common.Hex2BytesFixed(keyGenInfo.GroupIdHex, 32)
	var groupId [32]byte
	copy(groupId[:], groupIdRaw)
	myIndex := big.NewInt(int64(partyInfo.Index))

	// Locally store the generated public key
	pk := storage.GeneratedPubKeyInfo{
		PubKeyHashHex: pubKeyHash.Hex(),
		PubKeyHex:     result.Result,
		GroupIdHex:    keyGenInfo.GroupIdHex,
	}
	err = k.StoreGeneratedPubKeyInfo(&pk)
	if err != nil {
		return errors.Wrapf(err, "failed to store generated public key")
	}

	// Report the generated public key, in denormalized format due to Ethereum compatibility
	// Todo: establish a strategy to deal with "insufficient fund" error, maybe check account balance before report
	tx, err := k.ReportGeneratedKey(k.signer, groupId, myIndex, dnmGenPubKeyBytes)
	if err != nil {
		k.Error("Failed to report public key", logger.Field{"error", err})
		return errors.Wrap(err, "failed to report generated key")
	}

	// Locally update keygen request information
	keyGenInfo.PubKeyReportedAt = time.Now()
	keyGenInfo.PubKeyHashHex = pubKeyHash.Hex()
	err = k.StoreKeygenRequestInfo(keyGenInfo)
	if err != nil {
		return errors.WithStack(err)
	}

	k.pendingReports[tx.Hash()] = &ReportKeyTx{
		groupId:            groupId,
		myIndex:            myIndex,
		generatedPublicKey: dnmGenPubKeyBytes,
	}
	delete(k.pendingKeygenRequests, requestId)

	addr, _ := myCrypto.PubKeyHexToAddress(result.Result) // for debug
	k.Info("Generated and reported public key", []logger.Field{
		{"ethAddress", addr},
		{"generatedPubkey", result.Result},
		{"reportedPubkey", common.Bytes2Hex(dnmGenPubKeyBytes)}}...)
	return nil
}
