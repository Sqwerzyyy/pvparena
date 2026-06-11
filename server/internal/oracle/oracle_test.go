package oracle

import (
	"testing"
)

func TestGetBTCPrice(t *testing.T) {
	o := NewOracle()

	price, err := o.GetBTCPrice()
	if err != nil {
		t.Fatalf("failed to get BTC price: %v", err)
	}

	if price.Price <= 0 {
		t.Fatalf("expected positive price, got %f", price.Price)
	}

	t.Logf("BTC price: $%.2f at %s", price.Price, price.Timestamp.Format("15:04:05"))
}

func TestGetETHPrice(t *testing.T) {
	o := NewOracle()

	price, err := o.GetETHPrice()
	if err != nil {
		t.Fatalf("failed to get ETH price: %v", err)
	}

	if price.Price <= 0 {
		t.Fatalf("expected positive price, got %f", price.Price)
	}

	t.Logf("ETH price: $%.2f at %s", price.Price, price.Timestamp.Format("15:04:05"))
}

func TestGetMultipleAssets(t *testing.T) {
	o := NewOracle()

	assets := []string{"BTC", "ETH", "SOL"}

	for _, asset := range assets {
		price, err := o.GetPrice(asset)
		if err != nil {
			t.Fatalf("failed to get %s price: %v", asset, err)
		}
		t.Logf("%s price: $%.2f", price.Symbol, price.Price)
	}
}
