package polygon

import (
	"candles-api/data"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/go-resty/resty/v2"
	"time"
)

type Client struct {
	host   string
	apiKey string
}

func NewClient(host string, apiKey string) *Client {
	return &Client{host: host, apiKey: apiKey}
}

func (c *Client) GetLatestCandles(symbol string) []*data.Candle {
	client := resty.New()
	to := time.Now().Format(time.DateOnly)
	from := time.Now().Add(time.Hour * -24 * 7).Format(time.DateOnly)
	url := fmt.Sprintf(
		"https://%s/v2/aggs/ticker/C:%s/range/1/minute/%s/%s?adjusted=true&sort=asc&apiKey=%s",
		c.host, symbol, from, to, c.apiKey,
	)
	resp, err := client.R().Get(url)
	candles := make([]*data.Candle, 0)
	if err != nil {
		log.Errorf("cannot get candles from polygon %v", err)
		return candles
	}
	if resp.StatusCode() != 200 {
		log.Errorf("cannot get candles from polygon %s", string(resp.Body()))
		return candles
	}
	res := struct {
		Ticker  string `json:"ticker"`
		Results []struct {
			Open      float64 `json:"o"`
			High      float64 `json:"h"`
			Low       float64 `json:"l"`
			Close     float64 `json:"c"`
			Timestamp int64   `json:"t"`
		} `json:"results"`
	}{}
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		log.Errorf("cannot get candles from polygon %v", err)
		return candles
	}
	for _, item := range res.Results {
		closingTimestamp := item.Timestamp
		openPrice := item.Open
		highPrice := item.High
		lowPrice := item.Low
		closePrice := item.Close
		volume := 0.0
		turnover := 0.0
		candles = append(candles, data.NewCandle(
			symbol,
			"",
			60,
			uint64(closingTimestamp),
			uint64(closingTimestamp-60000),
			openPrice,
			closePrice,
			highPrice,
			lowPrice,
			volume,
			turnover,
		))
	}
	return candles
}
