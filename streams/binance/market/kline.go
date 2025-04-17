package binancemarket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"

	binancews "cryptoquant.com/m/internal/binance/ws"
	"cryptoquant.com/m/utils"
)

func SubscribeKline(ctx context.Context, symbol, interval string, handlers []func(binancews.KlineDataStream) error) error {
	var streamData binancews.Stream[binancews.KlineDataStream]

	// Binance Futures WebSocket endpoint
	url := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s@kline_%s",
		strings.ToLower(symbol), interval)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to Binance Futures kline stream for %s", symbol)

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

func HandlerKline[T any, V any](
	ch chan T,
	extractor func(binancews.KlineDataStream) V,
	converter func(V) T,
) func(binancews.KlineDataStream) error {
	return func(kline binancews.KlineDataStream) error {
		value := extractor(kline)

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

func ExtractOpen(kline binancews.KlineDataStream) string {
	return kline.Kline.OpenPrice
}

func ExtractClose(kline binancews.KlineDataStream) string {
	return kline.Kline.ClosePrice
}

func ExtractHigh(kline binancews.KlineDataStream) string {
	return kline.Kline.HighPrice
}

func ExtractLow(kline binancews.KlineDataStream) string {
	return kline.Kline.LowPrice
}

func NewOpenHandler(ch chan float64) func(binancews.KlineDataStream) error {
	return HandlerKline(ch, ExtractOpen, utils.StringToFloat64)
}

func NewCloseHandler(ch chan float64) func(binancews.KlineDataStream) error {
	return HandlerKline(ch, ExtractClose, utils.StringToFloat64)
}

func NewHighHandler(ch chan float64) func(binancews.KlineDataStream) error {
	return HandlerKline(ch, ExtractHigh, utils.StringToFloat64)
}

func NewLowHandler(ch chan float64) func(binancews.KlineDataStream) error {
	return HandlerKline(ch, ExtractLow, utils.StringToFloat64)
}
