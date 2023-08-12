package oracle_test

import (
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/crypto"
	"github.com/bianyuanop/oraclevm/oracle"
)

func StockAggregateTest(t *testing.T) {
	collection := oracle.NewEntityCollection(time.Now().Unix(), 0, 0)

	n := 10
	entities := make([]oracle.Entity, n)
	publisher := crypto.EmptyPublicKey
	stockName := "Stock-1"

	for i := 1; i < 10; i++ {
		var stock oracle.Entity = oracle.NewStock(stockName, uint64(i*1000), publisher, time.Now().Unix())

		entities[i] = stock
	}

	collection.Entities = entities

	// collection.Entities[0] = entities
}
