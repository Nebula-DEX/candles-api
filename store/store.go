package store

import (
	"candles-api/bybit"
	"candles-api/data"
	"candles-api/polygon"
	"candles-api/twelve_data"
	"github.com/charmbracelet/log"
	"maps"
	"sort"
	"sync"
	"time"
)

type PriceSource string

const (
	Bybit      PriceSource = "bybit"
	Polygon    PriceSource = "polygon"
	TwelveData PriceSource = "twelve-data"
)

type Interval struct {
	Seconds   uint64
	Retention time.Duration
}

type Config struct {
	MarketId    string
	PriceSource PriceSource
	Symbol      string
	MicCode     string
}

type Store struct {
	candles          map[string]map[uint64]map[uint64]*data.Candle
	intervals        []*Interval
	config           []*Config
	twelveDataClient *twelve_data.Client
	polygonClient    *polygon.Client
	bybitClient      *bybit.Client
	candlesLock      sync.RWMutex
}

func NewStore(
	intervals []*Interval,
	config []*Config,
	twelveDataClient *twelve_data.Client,
	polygonClient *polygon.Client,
	bybitClient *bybit.Client,
) *Store {
	return &Store{
		intervals:        intervals,
		config:           config,
		twelveDataClient: twelveDataClient,
		polygonClient:    polygonClient,
		bybitClient:      bybitClient,
		candles:          map[string]map[uint64]map[uint64]*data.Candle{},
	}
}

func (s *Store) SaveCandle(candle *data.Candle) {
	s.candlesLock.Lock()
	defer s.candlesLock.Unlock()
	if s.candles[candle.MarketId] == nil {
		s.candles[candle.MarketId] = map[uint64]map[uint64]*data.Candle{}
	}
	if s.candles[candle.MarketId][candle.Interval] == nil {
		s.candles[candle.MarketId][candle.Interval] = map[uint64]*data.Candle{}
	}
	s.candles[candle.MarketId][candle.Interval][candle.ClosingTimestamp] = candle
}

func (s *Store) RemoveCandle(candle *data.Candle) {
	s.candlesLock.Lock()
	defer s.candlesLock.Unlock()
	if s.candles[candle.MarketId] != nil {
		if s.candles[candle.MarketId][candle.Interval] != nil {
			delete(s.candles[candle.MarketId][candle.Interval], candle.ClosingTimestamp)
		}
	}
}

func (s *Store) GetCandles(marketId string, interval uint64, fromTimestamp uint64, toTimestamp uint64) []*data.Candle {
	s.candlesLock.RLock()
	defer s.candlesLock.RUnlock()
	candles := make([]*data.Candle, 0)
	if s.candles[marketId] != nil {
		if s.candles[marketId][interval] != nil {
			candlesList := maps.Values(s.candles[marketId][interval])
			for candle := range candlesList {
				if candle.ClosingTimestamp >= fromTimestamp && (candle.ClosingTimestamp <= toTimestamp || toTimestamp == 0) {
					candles = append(candles, &data.Candle{
						Id:               candle.Id,
						Symbol:           candle.Symbol,
						MarketId:         candle.MarketId,
						Interval:         candle.Interval,
						ClosingTimestamp: candle.ClosingTimestamp,
						OpeningTimestamp: candle.OpeningTimestamp,
						Open:             candle.Open,
						Close:            candle.Close,
						High:             candle.High,
						Low:              candle.Low,
						Volume:           candle.Volume,
						Turnover:         candle.Turnover,
					})
				}
			}
		}
	}
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].ClosingTimestamp > candles[j].ClosingTimestamp
	})
	return candles
}

func (s *Store) ArchiveCandles() {
	go func() {
		for range time.NewTicker(time.Second).C {
			started := time.Now()
			for _, interval := range s.intervals {
				oldestTimestamp := time.Now().Add(-interval.Retention).UnixMilli()
				for _, config := range s.config {
					candles := s.GetCandles(config.MarketId, interval.Seconds, 1, 0)
					for _, candle := range candles {
						if int64(candle.ClosingTimestamp) < oldestTimestamp {
							s.RemoveCandle(candle)
						}
					}
				}
			}
			ended := time.Now()
			log.Infof("archive took %.5f milliseconds", float64(ended.UnixNano()-started.UnixNano())/1000000.0)
		}
	}()
}

func (s *Store) GetStartingTimestampsForInterval(interval uint64, first uint64, last uint64) []uint64 {
	firstMinute := time.UnixMilli(int64(first)).Minute()
	lastMinute := time.UnixMilli(int64(last)).Minute()
	remainderStart := firstMinute % int(interval/60)
	remainderEnd := lastMinute % int(interval/60)
	start := time.UnixMilli(int64(first)).Add(time.Minute * time.Duration(remainderStart*-1))
	end := time.UnixMilli(int64(last))
	if remainderEnd > 0 {
		end = end.Add(time.Minute * time.Duration(interval/60))
	}
	results := make([]uint64, 0)
	for ts := start.UnixMilli(); ts < end.UnixMilli(); ts += int64(interval) * 1000 {
		results = append(results, uint64(time.UnixMilli(ts).UnixMilli()))
	}
	return results
}

func (s *Store) AggregateCandles() {
	go func() {
		for range time.NewTicker(time.Second).C {
			started := time.Now()
			for _, config := range s.config {
				candles := s.GetCandles(config.MarketId, 60, 1, 0)
				if len(candles) == 0 {
					continue
				}
				firstTs := candles[len(candles)-1].ClosingTimestamp
				lastTs := candles[0].ClosingTimestamp
				for _, interval := range s.intervals {
					if interval.Seconds == 60 {
						continue
					}
					fromTimestamps := s.GetStartingTimestampsForInterval(interval.Seconds, firstTs, lastTs)
					intervalCandles := make([]*data.Candle, 0)
					for _, ts := range fromTimestamps {
						openingTimestamp := ts + 60000
						closingTimestamp := ts + (interval.Seconds * 1000)
						candles = s.GetCandles(config.MarketId, 60, openingTimestamp, closingTimestamp)
						sort.Slice(candles, func(i, j int) bool {
							return candles[i].ClosingTimestamp < candles[j].ClosingTimestamp
						})
						openPrice := 0.0
						highPrice := 0.0
						lowPrice := 9999999999.0
						closePrice := 0.0
						volume := 0.0
						turnover := 0.0
						for i, candle := range candles {
							if i == 0 {
								openPrice = candle.Open
							}
							if i == len(candles)-1 {
								closePrice = candle.Close
							}
							if candle.High > highPrice {
								highPrice = candle.High
							}
							if candle.Low < lowPrice {
								lowPrice = candle.Low
							}
							volume += candle.Volume
							turnover += candle.Turnover
						}
						if len(candles) > 0 {
							intervalCandles = append(intervalCandles, data.NewCandle(
								config.Symbol,
								config.MarketId,
								interval.Seconds,
								closingTimestamp,
								openingTimestamp,
								openPrice,
								closePrice,
								highPrice,
								lowPrice,
								volume,
								turnover,
							))
						}
					}
					for _, candle := range intervalCandles {
						s.SaveCandle(candle)
					}
				}
			}
			ended := time.Now()
			log.Infof("aggregation took %.5f milliseconds", float64(ended.UnixNano()-started.UnixNano())/1000000.0)
		}
	}()
}

func (s *Store) SyncCandles() {
	go func() {
		for range time.NewTicker(time.Second).C {
			oldestTimestamp := time.Now().Add(-s.intervals[0].Retention).UnixMilli()
			for _, config := range s.config {
				go func() {
					candles := make([]*data.Candle, 0)
					if config.PriceSource == Bybit {
						candles = s.bybitClient.GetLatestCandles(config.Symbol)
					} else if config.PriceSource == Polygon {
						candles = s.polygonClient.GetLatestCandles(config.Symbol)
					}
					for _, candle := range candles {
						candle.MarketId = config.MarketId
						if int64(candle.ClosingTimestamp) >= oldestTimestamp {
							s.SaveCandle(candle)
						}
					}
				}()
			}
		}
	}()
	go func() {
		for range time.NewTicker(time.Second * 15).C {
			//oldestTimestamp := time.Now().Add(-s.intervals[0].Retention).UnixMilli()
			for _, config := range s.config {
				go func() {
					candles := make([]*data.Candle, 0)
					if config.PriceSource == TwelveData {
						candles = s.twelveDataClient.GetLatestCandles(config.Symbol, config.MicCode)
					}
					for _, candle := range candles {
						candle.MarketId = config.MarketId
						//if int64(candle.ClosingTimestamp) >= oldestTimestamp {
						s.SaveCandle(candle)
						//}
					}
				}()
			}
		}
	}()
}
