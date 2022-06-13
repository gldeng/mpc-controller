package storage

import (
	"context"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding storage

type Storer interface {
	StorerStoreGroupInfo
	StorerLoadGroupInfo
	StorerLoadGroupInfos

	StorerStoreParticipantInfo
	StorerLoadParticipantInfo
	StorerLoadParticipantInfos

	StorerGetParticipantIndex
	StorerGetGroupIds
	StorerGetPubKeys

	StorerStoreGeneratedPubKeyInfo
	StorerLoadGeneratedPubKeyInfo
	StorerLoadGeneratedPubKeyInfos

	StorerGetPariticipantKeys

	StorerStoreKeygenRequestInfo
	StorerLoadKeygenRequestInfo

	Close() error
}

// Group

type StorerStoreGroupInfo interface {
	StoreGroupInfo(ctx context.Context, g *GroupInfo) error
}

type StorerLoadGroupInfo interface {
	LoadGroupInfo(ctx context.Context, groupIdHex string) (*GroupInfo, error)
}

type StorerLoadGroupInfos interface {
	LoadGroupInfos(ctx context.Context) ([]*GroupInfo, error)
}

// Participant

type StorerStoreParticipantInfo interface {
	StoreParticipantInfo(ctx context.Context, p *ParticipantInfo) error
}

type StorerLoadParticipantInfo interface {
	LoadParticipantInfo(ctx context.Context, pubKeyHashHex, groupId string) (*ParticipantInfo, error)
}

type StorerLoadParticipantInfos interface {
	LoadParticipantInfos(ctx context.Context, pubKeyHashHex string) ([]*ParticipantInfo, error)
}

type StorerGetParticipantIndex interface {
	GetIndex(ctx context.Context, partiPubKeyHashHex, genPubKeyHexHex string) (*big.Int, error)
}

type StorerGetGroupIds interface {
	GetGroupIds(ctx context.Context, partiPubKeyHashHex string) ([][32]byte, error)
}

type StorerGetPubKeys interface {
	GetPubKeys(ctx context.Context, partiPubKeyHashHex string) ([][]byte, error)
}

// Generated public key

type StorerStoreGeneratedPubKeyInfo interface {
	StoreGeneratedPubKeyInfo(ctx context.Context, genPubKeyInfo *GeneratedPubKeyInfo) error
}

type StorerLoadGeneratedPubKeyInfo interface {
	LoadGeneratedPubKeyInfo(ctx context.Context, pubKeyHashHex string) (*GeneratedPubKeyInfo, error)
}

type StorerLoadGeneratedPubKeyInfos interface {
	LoadGeneratedPubKeyInfos(ctx context.Context, groupIdHexs []string) ([]*GeneratedPubKeyInfo, error)
}

type StorerGetPariticipantKeys interface {
	GetPariticipantKeys(ctx context.Context, genPubKeyHashHex string, indices []*big.Int) ([]string, error)
}

// Keygen request info

type StorerStoreKeygenRequestInfo interface {
	StoreKeygenRequestInfo(ctx context.Context, keygenReqInfo *KeygenRequestInfo) error
}

type StorerLoadKeygenRequestInfo interface {
	LoadKeygenRequestInfo(ctx context.Context, reqIdHex string) (*KeygenRequestInfo, error)
}
