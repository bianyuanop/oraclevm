package oracle_test

import (
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/crypto"
	"github.com/bianyuanop/oraclevm/oracle"
)

func TestStockAggregate(t *testing.T) {
	stockName := "Stock-1"
	collection := oracle.NewEntityCollection(time.Now().Unix(), 0, 0, stockName)

	n := 10
	entities := make([]oracle.Entity, n)
	publisher := crypto.EmptyPublicKey

	// 1000, 2000, ..., 10000
	for i := 1; i <= n; i++ {
		var stock oracle.Entity = oracle.NewStock(stockName, uint64(i*1000), publisher, time.Now().Unix())

		entities[i-1] = stock
	}

	collection.MergeMany(entities)

	r1, e1 := collection.Result()

	if e1 != nil {
		t.Errorf("error aggregation: %+v, %+v", e1, r1)
	}

	collection.RemoveMany(5)

	r2, e2 := collection.Result()

	// 6000, ..., 10000
	priceShouldBe := 8000

	if e2 != nil || priceShouldBe != int(r2.(*oracle.Stock).Price) {
		t.Errorf("error aggregation: %+v, %+v", e2, r2)
	}
}

func TestStockMarshal(t *testing.T) {
	payload := `{ "ticker": "Apple", "price": 1999 }`
	_, err := oracle.UnmarshalStock([]byte(payload))

	if err != nil {
		t.Errorf("Unmarshal stock failed: %+v\n", err)
	}

	// t.Errorf("%+v", stock)
}
