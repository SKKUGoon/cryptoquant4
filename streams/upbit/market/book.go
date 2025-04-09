package upbitmarket

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	upbitws "cryptoquant.com/m/internal/upbit/ws"
	"github.com/gorilla/websocket"
)

func SubscribeBook(ctx context.Context, symbol string, handlers []func(upbitws.SpotOrderbook) error) error {
	// Add random delay before initial connection attempt
	initialDelay := time.Duration(rand.Int63n(int64(maxInitialDelay)))
	log.Printf("Waiting %v before initial connection attempt for %s", initialDelay, symbol)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(initialDelay):
		// Continue with connection attempt
	}

	// Upbit Websocket endpoint
	url := "wss://api.upbit.com/websocket/v1"

	var conn *websocket.Conn
	var err error
	backoff := initialBackoff

	for retry := range maxRetries {
		if conn, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
			break
		}

		log.Printf("WebSocket connection attempt %d/%d failed: %v", retry+1, maxRetries, err)

		backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			continue
		}
	}

	if err != nil {
		return fmt.Errorf("websocket connection failed after %d retries: %v", maxRetries, err)
	}
	defer conn.Close()

	log.Printf("Connected to Upbit book stream for %s", symbol)

	subscriptionMessage := upbitws.SubscriptionMessage{
		upbitws.SubscriptionMessageTicket{
			Ticket: "book_subscription_" + time.Now().Format("20060102") + "_" + symbol,
		},
		upbitws.SubscriptionMessageType{
			Type:  "orderbook",
			Codes: []string{symbol},
			Level: 0,
		},
		upbitws.SubscriptionMessageFormat{
			Format: "DEFAULT",
		},
	}

	if err := conn.WriteJSON(subscriptionMessage); err != nil {
		return fmt.Errorf("failed to send subscription message: %v", err)
	}

	pingTicker := time.NewTicker(40 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("Failed to send ping: %v", err)
					_ = conn.Close()
					return
				} else {
					log.Printf("Ping sent")
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var orderbook upbitws.SpotOrderbook
			if err := conn.ReadJSON(&orderbook); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket closed unexpectedly: %v", err)
					// Try to reconnect
					return SubscribeBook(ctx, symbol, handlers)
				}
				log.Printf("Error reading message: %v", err)
				continue
			}
			for _, handler := range handlers {
				if err := handler(orderbook); err != nil {
					return fmt.Errorf("handler error: %v", err)
				}
			}
		}
	}
}

func HandlerBook[T any, V any](
	ch chan T,
	extractor func(upbitws.SpotOrderbook) V,
	converter func(V) T,
) func(upbitws.SpotOrderbook) error {
	return func(orderbook upbitws.SpotOrderbook) error {
		value := extractor(orderbook)

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

func ExtractBestBidPrc(orderbook upbitws.SpotOrderbook) float64 {
	return orderbook.OrderbookUnits[0].BidPrice
}

func ExtractBestBidQty(orderbook upbitws.SpotOrderbook) float64 {
	return orderbook.OrderbookUnits[0].BidSize
}

func ExtractBestAskPrc(orderbook upbitws.SpotOrderbook) float64 {
	return orderbook.OrderbookUnits[0].AskPrice
}

func ExtractBestAskQty(orderbook upbitws.SpotOrderbook) float64 {
	return orderbook.OrderbookUnits[0].AskSize
}

func NewBestBidPrcHandler(ch chan float64) func(upbitws.SpotOrderbook) error {
	return HandlerBook(ch, ExtractBestBidPrc, nil)
}

func NewBestBidQtyHandler(ch chan float64) func(upbitws.SpotOrderbook) error {
	return HandlerBook(ch, ExtractBestBidQty, nil)
}

func NewBestAskPrcHandler(ch chan float64) func(upbitws.SpotOrderbook) error {
	return HandlerBook(ch, ExtractBestAskPrc, nil)
}

func NewBestAskQtyHandler(ch chan float64) func(upbitws.SpotOrderbook) error {
	return HandlerBook(ch, ExtractBestAskQty, nil)
}
