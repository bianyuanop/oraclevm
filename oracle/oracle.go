package oracle

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/crypto"
)

const (
	StockID = 0
	SportID = iota
)

type Entity struct {
	ID        ids.ID `json:"id"`
	Publisher string `json:"string"`
	Tick      uint64 `json:"tick"`

	EntityType string `json:"type"`
	Payload    []byte `json:"payload"`

	_type     uint64
	publisher crypto.PublicKey
}

type Oracle struct {
	c Controller

	oracles map[uint64][]Entity
}
