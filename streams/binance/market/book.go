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

func SubscribeBook(ctx context.Context, symbol string, handlers []func(binancews.FutureBookTicker) error) error {
	var streamData binancews.Stream[binancews.FutureBookTicker]

	// Binance Futures Websocket endpoint
	url := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s@bookTicker", strings.ToLower(symbol))
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to Binance Futures book stream for %s", symbol)

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

// HandlerBook is a generic handler that inserts values into a channel
// T is the type of value to be inserted into the channel
// V is the type of value to be extracted from the trade
func HandlerBook[T any, V any](
	ch chan T,
	extractor func(binancews.FutureBookTicker) V,
	converter func(V) T,
) func(binancews.FutureBookTicker) error {
	return func(book binancews.FutureBookTicker) error {
		value := extractor(book)

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

func ExtractOrderbook(book binancews.FutureBookTicker) [2][2]float64 {
	return [2][2]float64{
		{utils.StringToFloat64(book.BestBidPrice), utils.StringToFloat64(book.BestBidQty)},
		{utils.StringToFloat64(book.BestAskPrice), utils.StringToFloat64(book.BestAskQty)},
	}
}

func NewOrderbookHandler(ch chan [2][2]float64) func(binancews.FutureBookTicker) error {
	return HandlerBook(ch, ExtractOrderbook, nil)
}
