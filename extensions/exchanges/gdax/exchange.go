package gdax

import (
	"net/http"
	"strconv"

	"github.com/cpurta/tatanka/internal/model"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
)

var (
	_ model.Exchange = &gdaxExchange{}
)

type gdaxExchange struct {
	client            *coinbasepro.Client
	tradeCursor       *coinbasepro.Cursor
	historyScan       string
	makerFee          float64
	takerFee          float64
	BackfillRateLimit int64
}

// NewGDAXExchange returns a new Coinbase(gdax) Exchange interface implementation
func NewGDAXExchange(key, passphrase, secret string, httpClient *http.Client) *gdaxExchange {
	client := coinbasepro.NewClient()

	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    "https://api.pro.coinbase.com",
		Key:        key,
		Passphrase: passphrase,
		Secret:     secret,
	})

	client.HTTPClient = httpClient

	return &gdaxExchange{
		client:            client,
		historyScan:       "backward",
		makerFee:          0.0,
		takerFee:          0.3,
		BackfillRateLimit: int64(335),
	}
}

// GetTrades returns the transaction history for a specific product on the Coinbase(gdax) exchange
func (exchange *gdaxExchange) GetTrades(productID string) ([]*model.Trade, error) {
	var (
		gdaxTrades []coinbasepro.Trade
		trades     = make([]*model.Trade, 0)
	)

	if exchange.tradeCursor == nil {
		exchange.tradeCursor = exchange.client.ListTrades(productID)
	}

	if !exchange.tradeCursor.HasMore {
		return trades, nil
	}

	if err := exchange.tradeCursor.NextPage(&gdaxTrades); err != nil {
		return nil, err
	}

	for _, trade := range gdaxTrades {
		size, _ := strconv.ParseFloat(trade.Size, 64)

		price, _ := strconv.ParseFloat(trade.Price, 64)

		trades = append(trades, &model.Trade{
			TradeID: strconv.Itoa(trade.TradeID),
			Size:    size,
			Price:   price,
			Time:    trade.Time.Time(),
			Side:    trade.Side,
		})
	}

	return trades, nil
}

// GetBalance returns the current account balance held on Coinbase(gdax)
func (exchange *gdaxExchange) GetBalance(currency string, asset string) (*model.Balance, error) {
	var (
		accounts []coinbasepro.Account
		balance  = &model.Balance{}
		err      error
	)

	if accounts, err = exchange.client.GetAccounts(); err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.Currency == currency {
			balance.Currency, _ = strconv.ParseFloat(account.Balance, 64)
			balance.CurrencyHold, _ = strconv.ParseFloat(account.Hold, 64)
		}

		if account.Currency == asset {
			balance.Asset, _ = strconv.ParseFloat(account.Balance, 64)
			balance.AssetHold, _ = strconv.ParseFloat(account.Hold, 64)
		}
	}

	return balance, nil
}

// GetQuote returns the current quote price of a product on Coinbase (gdax)
func (exchange *gdaxExchange) GetQuote(productID string) (*model.Quote, error) {
	var (
		ticker coinbasepro.Ticker
		quote  = &model.Quote{}
		err    error
	)

	if ticker, err = exchange.client.GetTicker(productID); err != nil {
		return nil, err
	}

	quote.Bid, _ = strconv.ParseFloat(ticker.Bid, 64)
	quote.Ask, _ = strconv.ParseFloat(ticker.Ask, 64)

	return quote, nil
}
