package binancemarket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	binancews "cryptoquant.com/m/internal/binance/ws"
	"cryptoquant.com/m/utils"
	"github.com/gorilla/websocket"
)

func SubscribeAggtrade(ctx context.Context, symbol string, handlers []func(binancews.FutureAggTrade) error) error {
	var streamData binancews.Stream[binancews.FutureAggTrade]

	// Binance Futures Websocket endpoint
	url := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s@aggTrade", strings.ToLower(symbol))
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to Binance Futures aggtrade stream for %s", symbol)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("error reading message: %v", err)
			}

			if err := json.Unmarshal(message, &streamData); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			for _, handler := range handlers {
				if err := handler(streamData.Data); err != nil {
					return fmt.Errorf("handler error: %v", err)
				}
			}
		}
	}
}

// HandlerAggtrade is a generic handler that inserts values into a channel
// T is the type of value to be inserted into the channel
// V is the type of value to be extracted from the trade
func HandlerAggtrade[T any, V any](
	ch chan T,
	extractor func(binancews.FutureAggTrade) V,
	converter func(V) T,
) func(binancews.FutureAggTrade) error {
	return func(trade binancews.FutureAggTrade) error {
		value := extractor(trade)

		if converter != nil {
			ch <- converter(value)
		} else {
			// Use type assertion to ensure type safety
			if v, ok := any(value).(T); ok {
				ch <- v
			} else {
				return fmt.Errorf("type mismatch: cannot convert %T to %T", value, *new(T))
			}
		}

		return nil
	}
}

// Example extractors for different fields
func ExtractPrice(trade binancews.FutureAggTrade) string {
	return trade.Price
}

func ExtractSymbol(trade binancews.FutureAggTrade) string {
	return trade.Symbol
}

func ExtractQuantity(trade binancews.FutureAggTrade) string {
	return trade.Quantity
}

func ExtractTrade(trade binancews.FutureAggTrade) [2]float64 {
	return [2]float64{utils.StringToFloat64(trade.Price), utils.StringToFloat64(trade.Quantity)}
}

// Helper functions for common use cases
func NewPriceHandler(ch chan float64) func(binancews.FutureAggTrade) error {
	return HandlerAggtrade(ch, ExtractPrice, utils.StringToFloat64)
}

func NewSymbolHandler(ch chan string) func(binancews.FutureAggTrade) error {
	return HandlerAggtrade(ch, ExtractSymbol, nil) // No converter needed since types match
}

func NewQuantityHandler(ch chan float64) func(binancews.FutureAggTrade) error {
	return HandlerAggtrade(ch, ExtractQuantity, utils.StringToFloat64)
}

func NewTradeHandler(ch chan [2]float64) func(binancews.FutureAggTrade) error {
	return HandlerAggtrade(ch, ExtractTrade, nil)
}

// Example usage:
// priceChan := make(chan float64)
// handler := NewPriceHandler(priceChan)
// err := SubscribeAggtrade(ctx, "BTCUSDT", handler)
