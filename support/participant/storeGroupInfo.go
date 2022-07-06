package participant

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// Trigger event: *contract.MpcManagerParticipantAdded
// Emit event: *events.GroupInfoStoredEvent

type GroupInfoStorer struct {
	Caller       bind.ContractCaller
	ContractAddr common.Address
	Logger       logger.Logger
	Publisher    dispatcher.Publisher
	Storer       storage.MarshalSetter
}

func (g *GroupInfoStorer) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerParticipantAdded:
		pubKeys, t, err := g.getGroupData(evt.GroupId)
		if err != nil {
			g.Logger.Error("Fail to get groupData", []logger.Field{{"error", err}, {"groupID", bytes.Bytes32ToHex(evt.GroupId)}}...)
			break
		}

		groupInfo := GroupInfo{
			GroupIdHex:     bytes.Bytes32ToHex(evt.GroupId),
			PartPubKeyHexs: pubKeys,
			Threshold:      t,
		}

		key, err := g.storeGroupInfo(evtObj.Context, &groupInfo)
		if err != nil {
			g.Logger.Error("Fail to store groupInfo", []logger.Field{{"error", err}, {"groupInfo", &groupInfo}}...)
			break
		}
		g.publishStoredEvent(evtObj.Context, key, &groupInfo, evtObj)
	}
}

func (g *GroupInfoStorer) getGroupData(groupID [32]byte) (partiPubKeyHexArr []string, threshold uint64, err error) {
	caller, err := contract.NewMpcManagerCaller(g.ContractAddr, g.Caller)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to create MpcManagerCaller")
	}

	groupData, err := caller.GetGroup(nil, groupID)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to get group data")
	}

	for _, k := range groupData.Participants {
		pk := bytes.BytesToHex(k)
		partiPubKeyHexArr = append(partiPubKeyHexArr, pk)
	}

	threshold = groupData.Threshold.Uint64()
	return
}

func (g *GroupInfoStorer) storeGroupInfo(ctx context.Context, groupInfo *GroupInfo) (key string, err error) {
	key = prefixGroupInfo + "-" + groupInfo.GroupIdHex
	err = g.Storer.MarshalSet(ctx, []byte(key), groupInfo)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return
}

func (g *GroupInfoStorer) publishStoredEvent(ctx context.Context, key string, groupInfo *GroupInfo, parentEvtObj *dispatcher.EventObject) {
	val := events.GroupInfo(*groupInfo)
	newEvt := events.GroupInfoStoredEvent{
		Key: key,
		Val: val,
	}

	g.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(parentEvtObj, "GroupInfoStorer", &newEvt, parentEvtObj.Context))
}
