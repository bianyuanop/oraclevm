package actions

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
)

const WarpQuerySize = consts.Uint64Len + consts.IDLen +
	consts.Uint64Len /* op bits */

type WarpQuery struct {
	EntityIndex        uint64 `json:"entityIndex"`
	DestinationChainID ids.ID `json:"destinationChainID"`
}

func (w *WarpQuery) Marshal() ([]byte, error) {
	p := codec.NewWriter(WarpQuerySize, WarpQuerySize)

	p.PackUint64(w.EntityIndex)
	p.PackID(w.DestinationChainID)

	return p.Bytes(), p.Err()
}

func UnmarshalWarpQuery(b []byte) (*WarpQuery, error) {
	var query WarpQuery
	p := codec.NewReader(b, WarpQuerySize)
	query.EntityIndex = p.UnpackUint64(false)
	p.UnpackID(true, &query.DestinationChainID)

	return &query, nil
}
