package polygon

import (
	"os"
	"testing"
)

var ApiKey = os.Getenv("POLYGON_API_KEY")

func TestClient_GetLatestCandles(t *testing.T) {
	client := NewClient("api.polygon.io", ApiKey)
	client.GetLatestCandles("EUR-USD")
}
