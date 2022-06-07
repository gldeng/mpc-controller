package storage

import "context"

type Storer interface {
	StoreGroupInfo(ctx context.Context, g *GroupInfo) error
	LoadGroupInfo(ctx context.Context, groupIdHex string) (*GroupInfo, error)
	LoadGroupInfos(ctx context.Context) ([]*GroupInfo, error)

	StoreParticipantInfo(ctx context.Context, p *ParticipantInfo) error
	LoadParticipantInfo(ctx context.Context, pubKeyHashHex, groupId string) (*ParticipantInfo, error)
	LoadParticipantInfos(ctx context.Context, pubKeyHashHex string) ([]*ParticipantInfo, error)

	StoreGeneratedPubKeyInfo(ctx context.Context, genPubKeyInfo *GeneratedPubKeyInfo) error
	LoadGeneratedPubKeyInfo(ctx context.Context, pubKeyHashHex string) (*GeneratedPubKeyInfo, error)
	LoadGeneratedPubKeyInfos(ctx context.Context, groupIdHexs []string) ([]*GeneratedPubKeyInfo, error)

	StoreKeygenRequestInfo(ctx context.Context, keygenReqInfo *KeygenRequestInfo) error
	LoadKeygenRequestInfo(ctx context.Context, reqIdHex string) (*KeygenRequestInfo, error)

	Close() error
}
