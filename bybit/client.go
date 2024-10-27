package bybit

import (
	"candles-api/data"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/go-resty/resty/v2"
	"strconv"
)

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{host: host}
}

func (c *Client) GetLatestCandles(symbol string) []*data.Candle {
	client := resty.New()
	url := fmt.Sprintf("https://%s/v5/market/kline?symbol=%s&interval=1&category=linear&limit=1000", c.host, symbol)
	resp, err := client.R().Get(url)
	candles := make([]*data.Candle, 0)
	if err != nil {
		log.Errorf("cannot get candles from bybit %v", err)
		return candles
	}
	if resp.StatusCode() != 200 {
		log.Errorf("cannot get candles from bybit %s", string(resp.Body()))
		return candles
	}
	res := struct {
		RetMsg string `json:"retMsg"`
		Result struct {
			Symbol string     `json:"symbol"`
			List   [][]string `json:"list"`
		}
	}{}
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		log.Errorf("cannot get candles from bybit %v", err)
		return candles
	}
	if res.RetMsg != "OK" {
		log.Errorf("cannot get candles from bybit %s", string(resp.Body()))
		return candles
	}
	for _, item := range res.Result.List {
		closingTimestamp, _ := strconv.ParseInt(item[0], 10, 0)
		openPrice, _ := strconv.ParseFloat(item[1], 64)
		highPrice, _ := strconv.ParseFloat(item[2], 64)
		lowPrice, _ := strconv.ParseFloat(item[3], 64)
		closePrice, _ := strconv.ParseFloat(item[4], 64)
		volume, _ := strconv.ParseFloat(item[5], 64)
		turnover, _ := strconv.ParseFloat(item[6], 64)
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
