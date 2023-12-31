package oracle_test

import (
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/crypto"
	"github.com/bianyuanop/oraclevm/oracle"
)

func TestEntityMetaMarshal(t *testing.T) {
	var entity oracle.Entity = oracle.NewStock("Test", 1000, crypto.EmptyPublicKey, time.Now().Unix())

	entityWithMeta := oracle.EntityWithMeta{
		ID:     0,
		Type:   oracle.StockID,
		Entity: entity,
	}

	marshalled := entityWithMeta.Marshal()

	restored, err := oracle.UnmarshalEntityWithMeta(marshalled)

	if err != nil {
		t.Error(err)
		t.Errorf("conversion failed: %+v\n%+v\n", marshalled, restored)
	}

	s1, _ := entityWithMeta.Entity.(*oracle.Stock)
	s2, _ := restored.Entity.(*oracle.Stock)

	if s1.Price != s2.Price && s1.Ticker != s2.Ticker && s1.Tick() != s2.Tick() && s1.Publisher() != s2.Publisher() {
		t.Errorf("value converted not equal: %+v is not equal to %+v\n", s1, s2)
	}
}

func TestUnmarshalEntity(t *testing.T) {
	payload := `{ "ticker": "Apple", "price": 1999 }`
	sType := oracle.StockID

	_, err := oracle.UnmarshalEntity(uint64(sType), []byte(payload))
	if err != nil {
		t.Errorf("Unmarshal entity failed: %+v\n", err)
	}

	// t.Errorf("%+v\n", entity)
}

func TestMarshalEntityMeta(t *testing.T) {
	payload := `{ "ticker": "Apple", "price": 1999 }`
	sType := oracle.StockID

	entity, err := oracle.UnmarshalEntity(uint64(sType), []byte(payload))
	if err != nil {
		t.Errorf("Unmarshal entity failed: %+v\n", err)
	}

	oracle.NewEntityWithMeta(uint64(sType), 0, entity)
}
