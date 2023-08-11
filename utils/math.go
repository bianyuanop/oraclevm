package utils

import (
	"errors"
)

var ErrInvalidPrice = errors.New("invalid price")
var ErrPriceParsing = errors.New("error parsing price")

// according to https://www.investopedia.com/terms/d/decimal-trading.asp
// the minimum decimal(increment) for stock price is $0.00001(for stock price below $1) or $0.001(for stock price higher than 1$)
// so we take 0.0001 as a price unit
func StockPriceToUint64(price float64) uint64 {
	return uint64(price * 10000)
}

func Uint64ToStockPrice(price uint64) float64 {
	var res float64
	res = float64(price)
	res /= 4

	return res
}
