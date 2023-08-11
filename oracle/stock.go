package oracle

type Stock struct {
	ticker string
	price  uint64
}

func (s *Stock) Aggregatable(stocks []*Stock) (res *Stock) {
	var priceSum uint64 = 0
	var counter uint64 = 1

	for _, stk := range stocks {
		if stk.ticker != s.ticker {
			continue
		}

		priceSum += stk.price
		counter += 1
	}

	res = new(Stock)

	res.ticker = s.ticker
	res.price = priceSum / counter

	return
}
