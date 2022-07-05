package cache

import "C"
import (
	"context"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"sync"
)

// Accept event: *events.GroupInfoStoredEvent
// Accept event: *events.ParticipantInfoStoredEvent
// Accept event: *events.GeneratedPubKeyInfoStoredEvent

type Cache struct {
	*sync.RWMutex
	GroupInfoMap           map[string]events.GroupInfo
	ParticipantInfoMap     map[string]events.ParticipantInfo
	GeneratedPubKeyInfoMap map[string]events.GeneratedPubKeyInfo
}

func (c *Cache) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.GroupInfoStoredEvent:
		c.Lock()
		c.GroupInfoMap[evt.Key] = evt.Val
		c.Unlock()
	case *events.ParticipantInfoStoredEvent:
		c.Lock()
		c.ParticipantInfoMap[evt.Key] = evt.Val
		c.Unlock()
	case *events.GeneratedPubKeyInfoStoredEvent:
		c.Lock()
		c.GeneratedPubKeyInfoMap[evt.Key] = evt.Val
		c.Unlock()
	}
}

func (c *Cache) LoadGroupInfo(groupIdHex string) *events.GroupInfo {
	c.RLock()
	defer c.RUnlock()

	key := events.PrefixGroupInfo + "-" + groupIdHex
	val := c.GroupInfoMap[key]
	return &val
}

func (c *Cache) GetMyIndex(myPubKeyHashHex, genPubKeyHashHex string) *big.Int {
	c.RLock()
	defer c.RUnlock()

	genPubKeyInfoStoredKey := events.PrefixGeneratedPubKeyInfo + "-" + genPubKeyHashHex
	groupIdHex := c.GeneratedPubKeyInfoMap[genPubKeyInfoStoredKey].GroupIdHex

	partiInfoStoredKey := events.PrefixParticipantInfo + "-" + myPubKeyHashHex + "-" + groupIdHex
	myIndex := c.ParticipantInfoMap[partiInfoStoredKey].Index

	return big.NewInt(int64(myIndex))
}

func (c *Cache) GetGeneratedPubKeyInfo(genPubKeyHashHex string) *events.GeneratedPubKeyInfo {
	c.RLock()
	defer c.RUnlock()

	genPubKeyInfoStoredKey := events.PrefixGeneratedPubKeyInfo + "-" + genPubKeyHashHex
	info := c.GeneratedPubKeyInfoMap[genPubKeyInfoStoredKey]
	return &info
}

func (c *Cache) GetParticipantKeys(genPubKeyHash common.Hash, indices []*big.Int) []string {
	pubKeyInfo := c.GetGeneratedPubKeyInfo(genPubKeyHash.Hex())
	if pubKeyInfo == nil {
		return nil
	}

	groupInfo := c.LoadGroupInfo(pubKeyInfo.GroupIdHex)
	if groupInfo == nil {
		return nil
	}

	var partPubKeyHexArr []string
	for _, ind := range indices {
		partPubKeyHex := groupInfo.PartPubKeyHexs[ind.Uint64()-1]
		partPubKeyHexArr = append(partPubKeyHexArr, partPubKeyHex)
	}
	return partPubKeyHexArr
}
