package cache

import "C"
import (
	"context"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"sync"
)

var _ ICache = (*Cache)(nil)

// Subscribe event: *events.GroupInfoStoredEvent
// Subscribe event: *events.ParticipantInfoStoredEvent
// Subscribe event: *events.GeneratedPubKeyInfoStoredEvent

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

func (c *Cache) GetNormalizedParticipantKeys(genPubKeyHash common.Hash, indices []*big.Int) ([]string, error) {
	partiKeys := c.GetParticipantKeys(genPubKeyHash, indices)
	if partiKeys == nil {
		return nil, nil
	}

	normalized, err := crypto.NormalizePubKeys(partiKeys)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return normalized, nil
}

func (c *Cache) IsParticipant(myPubKeyHash string, genPubKeyHash string, participantIndices []*big.Int) bool {
	myIndex := c.GetMyIndex(myPubKeyHash, genPubKeyHash)
	if myIndex == nil {
		return false
	}

	var participating bool
	for _, index := range participantIndices {
		if index.Cmp(myIndex) == 0 {
			participating = true
			break
		}
	}

	return participating
}
