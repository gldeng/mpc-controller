package storage

type Storer interface {
	StoreGroupInfo(g *GroupInfo) error
	LoadGroupInfo(groupIdHex string) (*GroupInfo, error)
	LoadGroupInfos() ([]*GroupInfo, error)

	StoreParticipantInfo(p *ParticipantInfo) error
	LoadParticipantInfo(pubKeyHashHex, groupId string) (*ParticipantInfo, error)
	LoadParticipantInfos(pubKeyHashHex string) ([]*ParticipantInfo, error)

	StoreGeneratedPubKeyInfo(genPubKeyInfo *GeneratedPubKeyInfo) error
	LoadGeneratedPubKeyInfo(pubKeyHashHex string) (*GeneratedPubKeyInfo, error)
	LoadGeneratedPubKeyInfos(groupIdHexs []string) ([]*GeneratedPubKeyInfo, error)

	StoreKeygenRequestInfo(keygenReqInfo *KeygenRequestInfo) error
	LoadKeygenRequestInfo(reqIdHex string) (*KeygenRequestInfo, error)

	Close() error
}
