package cache

import "C"
import (
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
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

func (c *Cache) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.GroupInfoStoredEvent:
		if len(c.GroupInfoMap) == 0 {
			c.GroupInfoMap = make(map[string]events.GroupInfo)
		}
		c.Lock()
		c.GroupInfoMap[evt.Key] = evt.Val
		c.Unlock()
	case *events.ParticipantInfoStoredEvent:
		if len(c.ParticipantInfoMap) == 0 {
			c.ParticipantInfoMap = make(map[string]events.ParticipantInfo)
		}
		c.Lock()
		c.ParticipantInfoMap[evt.Key] = evt.Val
		c.Unlock()
	case *events.GeneratedPubKeyInfoStoredEvent:
		if len(c.GeneratedPubKeyInfoMap) == 0 {
			c.GeneratedPubKeyInfoMap = make(map[string]events.GeneratedPubKeyInfo)
		}
		c.Lock()
		c.GeneratedPubKeyInfoMap[evt.Key] = evt.Val
		c.Unlock()
	}
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
