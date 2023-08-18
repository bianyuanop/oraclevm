package oracle_test

import (
	"testing"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/bianyuanop/oraclevm/oracle"
)

type Controller struct {
	logger logging.Logger
}

func (c *Controller) Logger() logging.Logger {
	return c.logger
}

func TestGetEntityCollectionMetas(t *testing.T) {
	controller := Controller{
		logger: logging.NoLog{},
	}

	trackedPairs := make([]string, 2)
	trackedPairs[0] = "Apple"
	trackedPairs[0] = "AMD"

	o := oracle.NewOracle(&controller, 0, trackedPairs)

	ecms := o.GetAvailableEntities()
	if len(ecms) != 2 {
		t.Error("Unexpected length of Entity Collection Metas")
	}

	apple := ecms[0]
	amd := ecms[0]

	if apple.EntityID != 0 && apple.EntityName != "Apple" && apple.EntityType != oracle.StockID {
		t.Errorf("Unexpected entity meta: %+v(real)", apple)
	}

	if amd.EntityID != 0 && amd.EntityName != "amd" && amd.EntityType != oracle.StockID {
		t.Errorf("Unexpected entity meta: %+v(real)", amd)
	}

}
