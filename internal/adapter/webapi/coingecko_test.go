package webapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoinGecko_symbolToCoinID(t *testing.T) {
	cg := &CoinGecko{}

	tests := []struct {
		symbol   string
		expected string
	}{
		{"BTC", "bitcoin"},
		{"btc", "bitcoin"},
		{"ETH", "ethereum"},
		{"USDT", "tether"},
		{"SOL", "solana"},
		{"XRP", "ripple"},
		{"BNB", "binancecoin"},
		{"DOGE", "dogecoin"},
		{"UNKNOWN", "unknown"}, // Неизвестная монета
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			result := cg.symbolToCoinID(tt.symbol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

