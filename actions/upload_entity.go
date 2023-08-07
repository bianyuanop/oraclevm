package actions

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/bianyuanop/oraclevm/storage"
)

type UploadEntity struct {
	Payload []byte `json:"payload"`

	EntityType  uint64 `json:"entity_type"`
	EntityIndex uint64 `json:"entity_index"`
}

func (*UploadEntity) GetTypeId() uint8 {
	return uploadEntityID
}

func (ue *UploadEntity) StateKeys(rauth chain.Auth, _ ids.ID) [][]byte {
	return [][]byte{
		storage.PrefixEntityKey(ue.EntityType, ue.EntityIndex),
	}
}

// func (ue *UploadEntity) Execute(
// 	ctx context.Context,
// 	r chain.Rules,
// 	db chain.Database,
// 	t int64,
// 	auth chain.Auth,
// 	txID ids.ID,
// 	wrapVerfied bool,
// ) (result *chain.Result, err error) {
// }

func (ue *UploadEntity) ValidRange(_ chain.Rules) (int64, int64) {
	return -1, -1
}

func (ue *UploadEntity) Marshal(p *codec.Packer) {
	p.PackUint64(ue.EntityType)
	p.PackUint64(ue.EntityIndex)

	p.PackBytes(ue.Payload)
}
