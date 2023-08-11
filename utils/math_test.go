package utils_test

import (
	"testing"

	"github.com/bianyuanop/oraclevm/utils"
)

func TestStockPriceNormalConversion(t *testing.T) {
	// 100.01
	price := 100.01
	price_restored := utils.StockPriceToUint64(price)

	if price_restored != 1000100 {
		t.Errorf("Conversion failed: %d", price_restored)
	}

}
