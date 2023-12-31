package oracle

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/bianyuanop/oraclevm/consts"
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

func UnmarshalEntity(_type uint64, payload []byte) (Entity, error) {
	switch int(_type) {
	case StockID:
		s, err := UnmarshalStock(payload)
		if err != nil {
			return nil, ErrMarshalEntityFailed
		}

		return Entity(s), nil
	default:
		return nil, ErrNotSupportedEntity
	}
}

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

type EntityCollectionMeta struct {
	EntityName string `json:"name"`
	EntityID   uint64 `json:"id"`
	EntityType uint64 `json:"type"`
}

type AggregationHistory struct {
	History []Entity
	Length  uint64
	l       sync.RWMutex
}

func NewAggregationHistory() *AggregationHistory {
	return &AggregationHistory{
		History: make([]Entity, 0),
		Length:  0,
	}
}

func (ah *AggregationHistory) GetHistory(limit uint64) []Entity {
	ah.l.RLock()
	defer ah.l.RUnlock()

	if limit > ah.Length {
		return ah.History[:ah.Length]
	}

	// latest limit ones
	return ah.History[ah.Length-limit:]
}

func (ah *AggregationHistory) Push(e Entity) {
	ah.l.Lock()
	defer ah.l.Unlock()

	// TODO: make it more efficient
	if ah.Length >= consts.HistoryCacheLen {
		lengthExceed := ah.Length - consts.HistoryCacheLen
		// truncate to cache len - 1
		ah.History = ah.History[lengthExceed+1:]
		ah.Length = consts.HistoryCacheLen - 1
	}

	ah.History = append(ah.History, e)
	ah.Length += 1
}

func (ah *AggregationHistory) Count() uint64 {
	ah.l.RLock()
	defer ah.l.RUnlock()

	return ah.Length
}

type Oracle struct {
	c Controller

	// _type -> EntityCollection
	oracles map[uint64]*EntityCollecton
	history map[uint64]*AggregationHistory
	counter uint64
}

func NewOracle(c Controller, t int64, trackedStocks []string) *Oracle {
	res := new(Oracle)

	res.c = c
	res.oracles = make(map[uint64]*EntityCollecton)
	res.history = make(map[uint64]*AggregationHistory)
	res.counter = 0

	sort.Strings(trackedStocks)

	res.counter = 0

	for _, ticker := range trackedStocks {
		res.oracles[res.counter] = NewEntityCollection(t, res.counter, StockID, ticker)
		res.history[res.counter] = NewAggregationHistory()

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

func (o *Oracle) ClearOracleNSaveHistory() {
	var i uint64
	for i = 0; i < o.counter; i++ {
		agg, err := o.oracles[i].Result()
		if err != nil {
			continue
		}
		o.history[i].Push(agg)
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

func (o *Oracle) GetEntityMeta(id uint64) (uint64, uint64, error) {
	if id > o.counter {
		return 0, 0, ErrOutOfEntityCollectionRange
	}

	return o.oracles[id].EntityID, o.oracles[id]._type, nil
}

func (o *Oracle) GetAggregatedResult(id uint64) (Entity, error) {
	if id > o.counter {
		return nil, ErrOutOfEntityCollectionRange
	}
	return o.oracles[id].Result()
}

func (o *Oracle) Counter() uint64 {
	return o.counter
}

func (o *Oracle) GetHistory(entityIndex uint64, limit uint64) ([]Entity, error) {
	if entityIndex >= o.counter {
		return make([]Entity, 0), ErrOutOfEntityCollectionRange
	}

	return o.history[entityIndex].GetHistory(limit), nil
}

func (o *Oracle) GetAvailableEntities() []*EntityCollectionMeta {
	ecms := make([]*EntityCollectionMeta, o.counter)

	var index uint64
	for index = 0; index < o.counter; index++ {
		ecms[index] = &EntityCollectionMeta{
			EntityName: o.oracles[index].EntityName,
			EntityID:   o.oracles[index].EntityID,
			EntityType: o.oracles[index]._type,
		}
	}

	return ecms
}

func (o *Oracle) GetEntityCollectionCount(index uint64) (uint64, error) {
	if index >= o.counter {
		return 0, ErrOutOfEntityCollectionRange
	}

	return o.history[index].Count(), nil
}
