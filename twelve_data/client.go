package twelve_data

import (
	"candles-api/data"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/go-resty/resty/v2"
	"strconv"
	"time"
)

type Client struct {
	host   string
	apiKey string
}

func NewClient(host string, apiKey string) *Client {
	return &Client{host: host, apiKey: apiKey}
}

func (c *Client) GetLatestCandles(symbol string, micCode string) []*data.Candle {
	client := resty.New()
	url := fmt.Sprintf(
		"https://%s/time_series?symbol=%s&interval=1min&apikey=%s&mic_code=%s",
		c.host, symbol, c.apiKey, micCode,
	)
	resp, err := client.R().Get(url)
	candles := make([]*data.Candle, 0)
	if err != nil {
		log.Errorf("cannot get candles from twelve data %v", err)
		return candles
	}
	if resp.StatusCode() != 200 {
		log.Errorf("cannot get candles from twelve data %s", string(resp.Body()))
		return candles
	}
	res := struct {
		Meta struct {
			Symbol string `json:"symbol"`
		} `json:"meta"`
		Values []struct {
			DateTime string `json:"datetime"`
			Open     string `json:"open"`
			High     string `json:"high"`
			Low      string `json:"low"`
			Close    string `json:"close"`
		} `json:"values"`
	}{}
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		log.Errorf("cannot get candles from twelve data %v", err)
		return candles
	}
	for _, item := range res.Values {
		dateTime, _ := time.Parse(time.DateTime, item.DateTime)
		hourDiff := dateTime.Sub(time.Now()).Hours()
		closingTimestamp := dateTime.Add(time.Duration(hourDiff*-1) * time.Hour).UnixMilli()
		openPrice, _ := strconv.ParseFloat(item.Open, 0)
		highPrice, _ := strconv.ParseFloat(item.High, 0)
		lowPrice, _ := strconv.ParseFloat(item.Low, 0)
		closePrice, _ := strconv.ParseFloat(item.Close, 0)
		volume := 0.0
		turnover := 0.0
		candles = append(candles, data.NewCandle(
			symbol,
			"",
			60,
			uint64(closingTimestamp),
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
