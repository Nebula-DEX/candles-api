package main

import (
	"candles-api/api"
	"candles-api/bybit"
	"candles-api/polygon"
	"candles-api/store"
	"candles-api/twelve_data"
	"os"
	"time"
)

var intervals = []*store.Interval{
	{
		Seconds:   60,
		Retention: time.Hour * 24,
	},
	{
		Seconds:   300,
		Retention: time.Hour * 24 * 3,
	},
	{
		Seconds:   900,
		Retention: time.Hour * 24 * 7,
	},
	{
		Seconds:   3600,
		Retention: time.Hour * 24 * 14,
	},
	{
		Seconds:   14400,
		Retention: time.Hour * 24 * 60,
	},
	{
		Seconds:   86400,
		Retention: time.Hour * 24 * 365,
	},
}

var config = []*store.Config{
	{
		MarketId:    "82b7c459a515e8404ca92fcfa3bef312d331abb2af40ae056de13c810a3c4c08",
		PriceSource: store.Polygon,
		Symbol:      "USD-JPY",
	},
	{
		MarketId:    "74711691b900bc8fea802ebb99d06c4ee326bda75058ac1c9637e9bc8233872d",
		PriceSource: store.Polygon,
		Symbol:      "GBP-USD",
	},
	{
		MarketId:    "c256ac0206dd6c4b2c443acd4590b156fc4f0f6963806780a374f1202cc68e85",
		PriceSource: store.Polygon,
		Symbol:      "USD-CNH",
	},
	{
		MarketId:    "778e7f4cd2414faf44d1e8a5391bbec87616aef5798bb2093f2db56704543c5f",
		PriceSource: store.Polygon,
		Symbol:      "EUR-USD",
	},
	{
		MarketId:    "d81a8bacb5e1a6b4bc8773d8af4e4ad29a5109e0ed4648ffe26c136c84cad3fc",
		PriceSource: store.Polygon,
		Symbol:      "AUD-USD",
	},
	{
		MarketId:    "6d8da2600e94db28a0ff024049d8c1fe1e6d26ba46fd3e43517f26c133caad93",
		PriceSource: store.Bybit,
		Symbol:      "BTCUSDT",
	},
	{
		MarketId:    "f4131d11f6294172a6f9526d1bf0eee832846a47e3f30a759948dfdb7659198a",
		PriceSource: store.Bybit,
		Symbol:      "ETHUSDT",
	},
	{
		MarketId:    "6d2e736f4b15a29f513db892bafbd3e93977222fe3a660241179e10665a7f574",
		PriceSource: store.Bybit,
		Symbol:      "SOLUSDT",
	},
	{
		MarketId:    "90cbdea8d4986173b2fbcbbec1fe7565e7fc1e3aa60b3ccb0e9d1a5a9eb18f19",
		PriceSource: store.TwelveData,
		Symbol:      "W_1",
		MicCode:     "COMMODITY",
	},
	{
		MarketId:    "95a8b0dcd0acdd6c0c0df61bb24283626abaeb2f66821173e13affbb076d2b76",
		PriceSource: store.TwelveData,
		Symbol:      "JO1",
		MicCode:     "COMMODITY",
	},
	{
		MarketId:    "f54044c1c87ff31509ea495d8bc55783864bbcd2ced04db8cd2ce64ef43d1f49",
		PriceSource: store.TwelveData,
		Symbol:      "LC1",
		MicCode:     "COMMODITY",
	},
	{
		MarketId:    "b47b9a2c8a9f69c01a54093ed81083f712ec88e98a0cc1358a621be3e8632116",
		PriceSource: store.TwelveData,
		Symbol:      "XAU/USD",
		MicCode:     "COMMODITY",
	},
	{
		MarketId:    "b0e849d267dc8b1e543a2109885b9f9dba600a733a3b30595e93e772862b6cb1",
		PriceSource: store.TwelveData,
		Symbol:      "NG/USD",
		MicCode:     "COMMODITY",
	},
	{
		MarketId:    "19fa4e7dcaf956efe33e5345bfd7a8ad3b4ea4634cdd12b3158321350f949009",
		PriceSource: store.TwelveData,
		Symbol:      "WTI/USD",
		MicCode:     "COMMODITY",
	},
	{
		MarketId:    "03d186c550ae6f13c1b0732320f1923c60767e37df5fa4099565a3db49691894",
		PriceSource: store.TwelveData,
		Symbol:      "FTSE",
	},
	{
		MarketId:    "a98b3eeea8bdc5afd0677869df89d9630a277a02f7336bbc4c074ce5f743b581",
		PriceSource: store.TwelveData,
		Symbol:      "GDAXI",
	},
	{
		MarketId:    "ee75df55c84dd341ce285fd65b7dc8f0857db977f6fb2875bce1beb405735a48",
		PriceSource: store.TwelveData,
		Symbol:      "N225",
	},
	{
		MarketId:    "2b851d11814da7e409ce6b0da8a62f0cf0e2fa4fb4a6344289aebbad1a79cb8d",
		PriceSource: store.TwelveData,
		Symbol:      "FCHI",
	},
}

func main() {
	twelveDataClient := twelve_data.NewClient("api.twelvedata.com", os.Getenv("TWELVE_DATA_API_KEY"))
	polygonClient := polygon.NewClient("api.polygon.io", os.Getenv("POLYGON_API_KEY"))
	bybitClient := bybit.NewClient("api.bybit.com")
	appStore := store.NewStore(intervals, config, twelveDataClient, polygonClient, bybitClient)
	appStore.SyncCandles()
	appStore.ArchiveCandles()
	appStore.AggregateCandles()
	restApi := api.NewApi(appStore)
	restApi.Start()
}
