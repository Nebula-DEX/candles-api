package twelve_data

import (
	"os"
	"testing"
)

var ApiKey = os.Getenv("TWELVE_DATA_API_KEY")

func TestClient_GetLatestCandles(t *testing.T) {
	client := NewClient("api.twelvedata.com", ApiKey)
	client.GetLatestCandles("WTI/USD", "COMMODITY")
}
