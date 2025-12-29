package webapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ICoinGecko interface {
	GetCurrentPrice(ctx context.Context, coinSymbol string) (float64, error)
}

type CoinGecko struct {
	client  *resty.Client
	baseURL string
}

type CoinGeckoSimplePriceResponse struct {
	Price map[string]float64 `json:"usd"`
}

func NewCoinGecko(client *resty.Client) *CoinGecko {
	return &CoinGecko{
		client:  client,
		baseURL: "https://api.coingecko.com/api/v3",
	}
}

// GetCurrentPrice получает текущую цену монеты в USD с retry логикой
func (c *CoinGecko) GetCurrentPrice(ctx context.Context, coinSymbol string) (float64, error) {
	// CoinGecko использует ID монет, а не символы
	// Нужно преобразовать символ в ID (например, BTC -> bitcoin, ETH -> ethereum)
	coinID := c.symbolToCoinID(coinSymbol)

	// Retry логика для обработки rate limiting
	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Экспоненциальная задержка: 2s, 4s, 8s
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}

		url := fmt.Sprintf("%s/simple/price", c.baseURL)

		var result map[string]map[string]float64

		resp, err := c.client.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"ids":           coinID,
				"vs_currencies": "usd",
			}).
			SetResult(&result).
			Get(url)

		if err != nil {
			return 0, fmt.Errorf("failed to get price from CoinGecko: %w", err)
		}

		// Если получили 429 (Too Many Requests), повторяем попытку
		if resp.StatusCode() == 429 {
			if attempt < maxRetries {
				continue
			}
			return 0, fmt.Errorf("CoinGecko rate limit exceeded after %d retries", maxRetries)
		}

		if resp.IsError() {
			return 0, fmt.Errorf("CoinGecko API error: status %d", resp.StatusCode())
		}

		// Извлекаем цену
		if priceData, ok := result[coinID]; ok {
			if price, ok := priceData["usd"]; ok {
				// Небольшая задержка между запросами для соблюдения rate limit
				time.Sleep(1 * time.Second)
				return price, nil
			}
		}

		return 0, fmt.Errorf("price not found for coin: %s (ID: %s)", coinSymbol, coinID)
	}

	return 0, fmt.Errorf("failed to get price after retries")
}

// symbolToCoinID преобразует символ монеты в CoinGecko ID
func (c *CoinGecko) symbolToCoinID(symbol string) string {
	// Приводим к нижнему регистру
	symbol = strings.ToLower(strings.TrimSpace(symbol))

	// Маппинг популярных монет
	mapping := map[string]string{
		"btc":   "bitcoin",
		"eth":   "ethereum",
		"usdt":  "tether",
		"bnb":   "binancecoin",
		"sol":   "solana",
		"xrp":   "ripple",
		"usdc":  "usd-coin",
		"ada":   "cardano",
		"avax":  "avalanche-2",
		"doge":  "dogecoin",
		"dot":   "polkadot",
		"matic": "matic-network",
		"link":  "chainlink",
		"uni":   "uniswap",
		"ltc":   "litecoin",
		"atom":  "cosmos",
		"etc":   "ethereum-classic",
		"xlm":   "stellar",
		"bch":   "bitcoin-cash",
		"near":  "near",
		"algo":  "algorand",
		"vet":   "vechain",
		"icp":   "internet-computer",
		"fil":   "filecoin",
		"apt":   "aptos",
		"hbar":  "hedera-hashgraph",
		"arb":   "arbitrum",
		"op":    "optimism",
		"ldo":   "lido-dao",
		"imx":   "immutable-x",
		"stx":   "blockstack",
		"inj":   "injective-protocol",
		"sui":   "sui",
		"sei":   "sei-network",
		"tia":   "celestia",
		"xvg":   "verge", // Verge
		"trx":   "tron",
		"shib":  "shiba-inu",
		"dai":   "dai",
		"wbtc":  "wrapped-bitcoin",
		"leo":   "leo-token",
		"ton":   "the-open-network",
		"okb":   "okb",
	}

	if coinID, ok := mapping[symbol]; ok {
		return coinID
	}

	// Если не нашли в маппинге, возвращаем символ как есть
	// (может сработать для некоторых монет)
	return symbol
}

