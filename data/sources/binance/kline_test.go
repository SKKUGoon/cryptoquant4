package binancesource_test

import (
	"testing"

	binancesource "cryptoquant.com/m/data/sources/binance"
)

func TestGetKlineData(t *testing.T) {
	bm := binancesource.NewBinanceFutureMarketData()
	t.Log(bm.GetStatus())
	// Check the weight usage and how it's filling up
	for range 10 {
		klineData, err := bm.GetKlineData("BTCUSDT", "1m", 100)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(klineData)
		t.Log(bm.GetStatus())
	}
}

func TestGetKlineClosePrices(t *testing.T) {
	bm := binancesource.NewBinanceFutureMarketData()
	klineData, err := bm.GetKlineData("BTCUSDT", "1m", 100)
	if err != nil {
		t.Fatal(err)
	}
	closePrices, err := klineData.GetKlineClosePrices()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(closePrices)
}

func TestGetKlineLatestCloseTime(t *testing.T) {
	bm := binancesource.NewBinanceFutureMarketData()
	klineData, err := bm.GetKlineData("BTCUSDT", "1m", 100)
	if err != nil {
		t.Fatal(err)
	}
	closeTime, err := klineData.GetKlineLatestCloseTime()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(closeTime)
}
