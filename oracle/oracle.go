package oracle

import (
	"github.com/ava-labs/hypersdk/crypto"
)

const (
	StockID = 0
	SportID = iota
)

type Entity struct {
	Publisher string `json:"string"`
	Tick      uint64 `json:"tick"`

	Payload []byte `json:"payload"`

	publisher crypto.PublicKey
}

// TODO: do we need lock?
type EntityCollecton struct {
	MinTick uint64
	MaxTick uint64

	EntityID   uint64
	EntityType string

	// Stock Index -> FIFO queue
	Entities map[uint64][]*Entity
	_type    uint64
}

type Oracle struct {
	c Controller

	// _type -> EntityCollection
	oracles map[uint64]*EntityCollecton
}
