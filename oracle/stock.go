package oracle

import (
	"encoding/json"
	"time"

	"github.com/ava-labs/hypersdk/crypto"
)

type Stock struct {
	Ticker string `json:"ticker"`
	Price  uint64 `json:"price"`

	publisher crypto.PublicKey
	tick      int64
}

func NewStock(ticker string, price uint64, publisher crypto.PublicKey, tick int64) (s *Stock) {
	s = new(Stock)
	s.Price = price
	s.Ticker = ticker
	s.publisher = publisher
	s.tick = tick

	return s
}

func (s *Stock) Publisher() string {
	return string(s.publisher[:])
}

func (s *Stock) Tick() int64 {
	return s.tick
}

type StockAggregator struct {
	ticker string
	sum    uint64
	count  uint64
}

func NewStockAggregator(name string) *StockAggregator {
	res := new(StockAggregator)
	res.count = 0
	res.sum = 0
	// empty at first
	res.ticker = ""

	return res
}

func (sa *StockAggregator) Result() (Entity, error) {
	res := new(Stock)
	res.Ticker = sa.ticker
	res.publisher = crypto.EmptyPublicKey
	res.tick = time.Now().Unix()

	if sa.count != 0 {
		res.Price = sa.sum / sa.count
		return res, nil
	} else {
		return nil, ErrZeroDenominator
	}
}

func (sa *StockAggregator) MergeOne(s Entity) {

	stk, ok := s.(*Stock)
	if !ok {
		return
	}

	if sa.ticker == "" {
		sa.ticker = stk.Ticker
	}
	sa.sum += stk.Price
	sa.count += 1
}

func (sa *StockAggregator) RemoveOne(s Entity) {
	stk, ok := s.(*Stock)
	if !ok {
		return
	}
	sa.sum -= stk.Price
	sa.count -= 1
}

func (s *Stock) Marshal() []byte {
	// should always success
	res, _ := json.Marshal(s)

	return res
}

func UnmarshalStock(payload []byte) (*Stock, error) {
	var s Stock
	err := json.Unmarshal(payload, &s)

	if err != nil {
		return nil, err
	}
	return &s, nil
}
