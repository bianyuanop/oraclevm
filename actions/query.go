package actions

import (
	"context"
	"encoding/json"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/bianyuanop/oraclevm/storage"
)

type QueryResult struct {
	EntityType uint64 `json:"entityType"`
	Payload    []byte `json:"payload"`
}

type Query struct {
	warpQuery   *WarpQuery
	warpMessage *warp.Message
}

func (*Query) GetTypeID() uint8 {
	return queryID
}

func (q *Query) StateKeys(rauth chain.Auth, _ ids.ID) [][]byte {
	keys := [][]byte{
		storage.PrefixAggregationCacheResult(q.warpQuery.EntityIndex),
	}

	return keys
}

func (q *Query) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	t int64,
	rauth chain.Auth,
	_ ids.ID,
	warpVerified bool,
) (*chain.Result, error) {
	// _ := auth.GetActor(rauth)
	unitsUsed := q.MaxUnits(r)
	if !warpVerified {
		return &chain.Result{
			Success: false,
			Units:   unitsUsed,
			Output:  OutputWarpVerificationFailed,
		}, nil
	}

	var queryRes QueryResult
	entityType, payload, err := storage.GetCachedAggregationResult(ctx, db, q.warpQuery.EntityIndex)
	if err != nil {
		return &chain.Result{
			Success: false,
			Units:   unitsUsed,
			Output:  OutputEntityNotRecorded,
		}, nil
	}

	queryRes.EntityType = entityType
	// payload is `Entity.Marshal()`
	queryRes.Payload = payload

	wmPayload, err := json.Marshal(queryRes)
	if err != nil {
		return &chain.Result{
			Success: false,
			Units:   unitsUsed,
			Output:  OutputQueryResMarshalFailed,
		}, nil
	}

	wm := &warp.UnsignedMessage{
		Payload: wmPayload,
	}

	return &chain.Result{Success: true, Units: unitsUsed, WarpMessage: wm}, nil
}

func (q *Query) MaxUnits(chain.Rules) uint64 {
	return uint64(len(q.warpMessage.Payload))
}

func (*Query) Size() int {
	return 0
}

func (q *Query) Marshal(_ *codec.Packer) {}

func UnmarshalQuery(p *codec.Packer, wm *warp.Message) (chain.Action, error) {
	var (
		query Query
		err   error
	)

	query.warpMessage = wm
	query.warpQuery, err = UnmarshalWarpQuery(query.warpMessage.Payload)

	if err != nil {
		return nil, err
	}

	return &query, nil
}

func (q *Query) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}
