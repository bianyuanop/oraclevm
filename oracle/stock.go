package oracle

import (
	"time"

	"github.com/ava-labs/hypersdk/crypto"
)

type Stock struct {
	Ticker string
	Price  uint64

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

func NewStockAggregator() *StockAggregator {
	res := new(StockAggregator)
	res.count = 0
	res.sum = 0
	// res.ticker =

	return res
}

func (sa *StockAggregator) Result() (*Stock, error) {
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

func (sa *StockAggregator) MergeOne(s *Stock) {
	sa.sum += s.Price
	sa.count += 1
}

func (sa *StockAggregator) RemoveOne(s *Stock) {
	sa.sum -= s.Price
	sa.count -= 1
}
