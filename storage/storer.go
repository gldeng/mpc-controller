package storage

import (
	"context"
	"math/big"
)

type Storer interface {
	StoreGroupInfo(ctx context.Context, g *GroupInfo) error
	LoadGroupInfo(ctx context.Context, groupIdHex string) (*GroupInfo, error)
	LoadGroupInfos(ctx context.Context) ([]*GroupInfo, error)

	StoreParticipantInfo(ctx context.Context, p *ParticipantInfo) error
	LoadParticipantInfo(ctx context.Context, pubKeyHashHex, groupId string) (*ParticipantInfo, error)
	LoadParticipantInfos(ctx context.Context, pubKeyHashHex string) ([]*ParticipantInfo, error)

	GetIndex(ctx context.Context, partiPubKeyHashHex, genPubKeyHexHex string) (*big.Int, error)
	GetGroupIds(ctx context.Context, partiPubKeyHashHex string) ([][32]byte, error)
	GetPubKeys(ctx context.Context, partiPubKeyHashHex string) ([][]byte, error)

	StoreGeneratedPubKeyInfo(ctx context.Context, genPubKeyInfo *GeneratedPubKeyInfo) error
	LoadGeneratedPubKeyInfo(ctx context.Context, pubKeyHashHex string) (*GeneratedPubKeyInfo, error)
	LoadGeneratedPubKeyInfos(ctx context.Context, groupIdHexs []string) ([]*GeneratedPubKeyInfo, error)

	GetPariticipantKeys(ctx context.Context, genPubKeyHashHex string, indices []*big.Int) ([]string, error)

	StoreKeygenRequestInfo(ctx context.Context, keygenReqInfo *KeygenRequestInfo) error
	LoadKeygenRequestInfo(ctx context.Context, reqIdHex string) (*KeygenRequestInfo, error)

	Close() error
}
