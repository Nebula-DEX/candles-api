package data

import "fmt"

type Candle struct {
	Id               string  `json:"id,omitempty"`
	Symbol           string  `json:"symbol,omitempty"`
	MarketId         string  `json:"marketId,omitempty"`
	Interval         uint64  `json:"interval,omitempty"`
	ClosingTimestamp uint64  `json:"closingTimestamp,omitempty"`
	OpeningTimestamp uint64  `json:"openingTimestamp,omitempty"`
	Open             float64 `json:"open,omitempty"`
	Close            float64 `json:"close,omitempty"`
	High             float64 `json:"high,omitempty"`
	Low              float64 `json:"low,omitempty"`
	Volume           float64 `json:"volume,omitempty"`
	Turnover         float64 `json:"turnover,omitempty"`
}

func NewCandle(
	symbol string,
	marketId string,
	interval uint64,
	closingTimestamp uint64,
	openingTimestamp uint64,
	open float64,
	close float64,
	high float64,
	low float64,
	volume float64,
	turnover float64,
) *Candle {
	return &Candle{
		Id:               fmt.Sprintf("%s_%d_%d", symbol, closingTimestamp, interval),
		Symbol:           symbol,
		MarketId:         marketId,
		Interval:         interval,
		ClosingTimestamp: closingTimestamp,
		OpeningTimestamp: openingTimestamp,
		Open:             open,
		Close:            close,
		High:             high,
		Low:              low,
		Volume:           volume,
		Turnover:         turnover,
	}
}
