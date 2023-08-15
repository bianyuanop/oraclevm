package storage_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/crypto"
	"github.com/bianyuanop/oraclevm/oracle"
	"github.com/bianyuanop/oraclevm/storage"
)

func TestPackEnitty(t *testing.T) {
	entityIndex := 0
	entityType := oracle.StockID
	tick := time.Now().Unix()
	publisher := crypto.EmptyPublicKey

	stock := oracle.NewStock("Apple", 10000, publisher, tick)
	payload := stock.Marshal()

	packed := storage.PackEntity(uint64(entityIndex), uint64(entityType), tick, publisher, payload)

	uI, uTtype, uTick, uPub, uPayload := storage.UnpackEntity(packed)
	if uI != uint64(entityIndex) || uTtype != uint64(entityType) || tick != uTick || publisher != uPub || reflect.DeepEqual(payload, uPayload) {
		t.Errorf("packed entity is not equal to unpacked Entity")
	}
}
