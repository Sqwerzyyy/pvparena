package oracle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PriceData struct {
	Symbol    string
	Price     float64
	Timestamp time.Time
}

type Oracle struct {
	client  *http.Client
	baseURL string
}

func NewOracle() *Oracle {
	return &Oracle{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: "https://api.binance.com",
	}
}

func (o *Oracle) GetPrice(symbol string) (*PriceData, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%sUSDT", o.baseURL, symbol)

	resp, err := o.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var price float64
	fmt.Sscanf(result.Price, "%f", &price)

	return &PriceData{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}

func (o *Oracle) GetBTCPrice() (*PriceData, error) {
	return o.GetPrice("BTC")
}

func (o *Oracle) GetETHPrice() (*PriceData, error) {
	return o.GetPrice("ETH")
}

func (o *Oracle) WatchPrice(symbol string, interval time.Duration, callback func(*PriceData)) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			price, err := o.GetPrice(symbol)
			if err != nil {
				fmt.Printf("oracle error: %v\n", err)
				continue
			}
			callback(price)
		}
	}()
}
