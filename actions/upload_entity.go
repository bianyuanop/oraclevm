package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"

	"github.com/bianyuanop/oraclevm/auth"
	"github.com/bianyuanop/oraclevm/consts"
	"github.com/bianyuanop/oraclevm/oracle"
	"github.com/bianyuanop/oraclevm/storage"
)

type UploadEntity struct {
	Payload []byte `json:"payload"`

	EntityType  uint64 `json:"entity_type"`
	EntityIndex uint64 `json:"entity_index"`
}

func (*UploadEntity) GetTypeID() uint8 {
	return uploadEntityID
}

func (ue *UploadEntity) StateKeys(rauth chain.Auth, _ ids.ID) [][]byte {
	return [][]byte{
		storage.PrefixEntityKey(ue.EntityType, ue.EntityIndex),
	}
}

func (ue *UploadEntity) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	t int64,
	rauth chain.Auth,
	txID ids.ID,
	wrapVerfied bool,
) (result *chain.Result, err error) {
	actor := auth.GetActor(rauth)
	unitsUsed := ue.MaxUnits(r)

	if len(ue.Payload) > consts.PayloadMaxLen {
		return &chain.Result{Success: false, Units: unitsUsed, Output: PayloadSizeTooLarge}, nil
	}

	// try marshal payload
	output, err := oracle.UnmarshalEntity(ue.EntityType, ue.Payload)
	if err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, err
	}

	if err := storage.StoreEntity(ctx, db, ue.EntityType, ue.EntityIndex, t, actor, ue.Payload); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, err
	}

	return &chain.Result{Success: true, Units: unitsUsed, Output: output.Marshal()}, nil
}

func (ue *UploadEntity) ValidRange(_ chain.Rules) (int64, int64) {
	return -1, -1
}

func (ue *UploadEntity) Marshal(p *codec.Packer) {
	p.PackUint64(ue.EntityType)
	p.PackUint64(ue.EntityIndex)

	p.PackBytes(ue.Payload)
}

func (ue *UploadEntity) MaxUnits(chain.Rules) uint64 {
	return uint64(len(ue.Payload))
}

func (ue *UploadEntity) Size() int {
	return len(ue.Payload) + hconsts.Uint64Len*2
}
