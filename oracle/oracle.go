package oracle

import (
	"github.com/ava-labs/hypersdk/crypto"
)

const (
	StockID = 0
	SportID = iota
)

func EntityIDToTypeString(id uint64) (res string) {
	switch id {
	case 0:
		res = "Stock"
	case 1:
		res = "Sports"
	default:
		res = "Unknown"
	}

	return
}

type Entity struct {
	Publisher string `json:"string"`
	Tick      int64  `json:"tick"`

	inner Aggregatable

	publisher crypto.PublicKey
}

// TODO: do we need lock?
type EntityCollecton struct {
	MinTick int64
	MaxTick int64

	EntityID   uint64
	EntityType string

	// Stock Index -> FIFO queue
	Entities map[uint64][]*Entity
	_type    uint64
}

func NewEntityCollection(t int64, id uint64, _type uint64) (ec *EntityCollecton) {
	ec = new(EntityCollecton)
	ec._type = _type
	ec.EntityID = id
	ec.MaxTick = t
	ec.MinTick = t

	ec.EntityType = EntityIDToTypeString(_type)
	ec.Entities = make(map[uint64][]*Entity)

	return
}

// func (ec *EntityCollecton) Insert(id uint64, entity *Entity) (bool, error) {
// 	ec.Entities[id]
// }

type Aggregatable interface {
	aggregate([]*Aggregatable) Aggregatable
}

type Oracle struct {
	c Controller

	// _type -> EntityCollection
	oracles map[uint64]*EntityCollecton
}
