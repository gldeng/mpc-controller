package network

import (
	"github.com/ava-labs/avalanchego/ids"
)

const (
	AVAXID = "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
)

func AssetIDAVAX() (*ids.ID, error) {
	id, err := ids.FromString(AVAXID)
	if err != nil {
		return nil, err
	}
	return &id, err
}
