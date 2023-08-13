package oracle

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

const (
	StockID = 0
	SportID = iota
)

func EntityIDToTypeString(id uint64) (res string) {
	switch id {
	case 0:
		// should return `*oracle.Stock`
		res = reflect.TypeOf(&Stock{}).String()
	default:
		res = "Unknown"
	}

	return
}

func EntityName(id uint64, _type uint64) (res string) {
	res = fmt.Sprintf("%d-%s", id, EntityIDToTypeString(_type))

	return
}

type EntityWithMeta struct {
	Type   uint64 `json:"type"`
	ID     uint64 `json:"id"`
	Entity Entity `json:"entity"`
}

func NewEntityWithMeta(_type uint64, id uint64, entity Entity) *EntityWithMeta {
	return &EntityWithMeta{
		Type:   _type,
		ID:     id,
		Entity: entity,
	}
}

func (ewm *EntityWithMeta) Marshal() []byte {
	res, _ := json.Marshal(ewm)

	return res
}

func UnmarshalEntityWithMeta(payload []byte) (*EntityWithMeta, error) {
	var res *EntityWithMeta = &EntityWithMeta{}

	// same with `EntityWIthMeta` except we defer value decoding
	var data struct {
		Type   uint64          `json:"type"`
		ID     uint64          `json:"id"`
		Entity json.RawMessage `json:"entity"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	res.ID = data.Type
	res.ID = data.ID

	for _, impl := range entityKnownImplementations {
		_type := reflect.TypeOf(impl)
		if _type.String() == EntityIDToTypeString(data.Type) {
			target := reflect.New(_type)
			if err := json.Unmarshal(data.Entity, target.Interface()); err != nil {
				return nil, err
			}

			res.Entity = target.Elem().Interface().(Entity)
			break
		}
	}

	return res, nil
}

type Entity interface {
	Publisher() string
	Tick() int64
	Marshal() []byte
}

// to be used for dynamically unmarshal and marshal EntityWithMeta
var (
	entityKnownImplementations = []Entity{
		&Stock{},
	}
)

type EntityAggregator interface {
	Result() (Entity, error)
	MergeOne(Entity)
	RemoveOne(Entity)
}

type DefaultAggregator struct{}

func NewDefaultAggregator() *DefaultAggregator {
	return &DefaultAggregator{}
}

func (da *DefaultAggregator) Result() (Entity, error) {
	return nil, nil
}
func (da *DefaultAggregator) MergeOne(Entity)  {}
func (da *DefaultAggregator) RemoveOne(Entity) {}

// type Aggregatable interface {
// 	Aggregate([]Entity) (Entity, error)
// }

// TODO: do we need lock?
type EntityCollecton struct {
	MinTick int64
	MaxTick int64

	EntityID   uint64
	EntityType string
	// e.g. stock ticker name
	EntityName string

	// FIFO queue
	Entities []Entity

	aggregator EntityAggregator
	_type      uint64
}

func AggregatorFactory(_type uint64, name string) (aggregator EntityAggregator) {
	switch _type {
	case 0:
		aggregator = NewStockAggregator(name)
	default:
		aggregator = NewDefaultAggregator()
	}

	return
}

func NewEntityCollection(t int64, id uint64, _type uint64, name string) (ec *EntityCollecton) {
	ec = new(EntityCollecton)
	ec._type = _type
	ec.EntityID = id
	ec.MaxTick = t
	ec.MinTick = t
	ec.EntityName = name

	ec.EntityType = EntityIDToTypeString(_type)
	ec.Entities = make([]Entity, 0)

	ec.aggregator = AggregatorFactory(_type, name)

	return
}

func (ec *EntityCollecton) Result() (Entity, error) {
	return ec.aggregator.Result()
}

func (ec *EntityCollecton) MergeMany(es []Entity) {
	ec.Entities = append(ec.Entities, es...)

	for _, e := range es {
		ec.aggregator.MergeOne(e)
	}
}

func (ec *EntityCollecton) RemoveMany(count int) {
	length := len(ec.Entities)

	var numRemove int

	if length > count {
		numRemove = count
	} else {
		numRemove = length
	}

	var x Entity

	for i := 0; i < numRemove; i++ {
		x, ec.Entities = ec.Entities[0], ec.Entities[1:]
		ec.aggregator.RemoveOne(x)
	}
}

func (ec *EntityCollecton) RemoveBeforeTick(t int64) {
	var x Entity
	for {
		if len(ec.Entities) == 0 {
			break
		}

		x, ec.Entities = ec.Entities[0], ec.Entities[1:]
		ec.aggregator.RemoveOne(x)

		if x.Tick() >= t {
			break
		}
	}
}

func (ec *EntityCollecton) Clear() {
	ec.Entities = make([]Entity, 0)
	ec.aggregator = AggregatorFactory(ec._type, ec.EntityName)
}

type Oracle struct {
	c Controller

	// _type -> EntityCollection
	oracles map[uint64]*EntityCollecton
	counter uint64
}

func NewOracle(c Controller, t int64, trackedStocks []string) *Oracle {
	res := new(Oracle)

	res.c = c
	res.oracles = make(map[uint64]*EntityCollecton)
	res.counter = 0

	sort.Strings(trackedStocks)

	res.counter = 0

	for _, ticker := range trackedStocks {
		res.oracles[res.counter] = NewEntityCollection(t, res.counter, StockID, ticker)

		res.counter += 1
	}

	return res
}

func (o *Oracle) ClearEntityCollection() {
	var i uint64
	for i = 0; i < o.counter; i++ {
		o.oracles[i].Clear()
	}
}

func (o *Oracle) InsertEntity(id uint64, _type uint64, e Entity) error {
	if id >= o.counter {
		return ErrOutOfEntityCollectionRange
	}

	if _type != o.oracles[id]._type {
		return ErrUnexpectedEntityType
	}

	o.oracles[id].MergeMany([]Entity{e})

	return nil
}

func (o *Oracle) GetAggregatedResult(id uint64) (Entity, error) {
	if id > o.counter {
		return nil, ErrOutOfEntityCollectionRange
	}
	return o.oracles[id].Result()
}

func UnmarshalEntity(_type uint64, payload []byte) (Entity, error) {
	switch _type {
	case SportID:
		s, _ := UnmarshalStock(payload)
		return Entity(s), ErrMarshalEntityFailed
	default:
		return nil, ErrNotSupportedEntity
	}
}
