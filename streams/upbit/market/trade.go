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

const (
	maxRetries     = 5
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
	// Add random delay between 0-5 seconds for initial connection
	maxInitialDelay = 5 * time.Second
)

func SubscribeTrade(ctx context.Context, symbol string, handlers []func(upbitws.SpotTrade) error) error {
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

	// Retry connection with exponential backoff
	for retry := 0; retry < maxRetries; retry++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			break
		}

		log.Printf("WebSocket connection attempt %d failed: %v", retry+1, err)

		// Calculate next backoff duration
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

	log.Printf("Connected to Upbit trade stream for %s", symbol)

	// Create subscription message with three segments
	subscriptionMessage := upbitws.SubscriptionMessage{
		upbitws.SubscriptionMessageTicket{
			Ticket: "trade_subscription_" + time.Now().Format("20060102") + "_" + symbol,
		},
		upbitws.SubscriptionMessageType{
			Type:  "trade",
			Codes: []string{symbol},
		},
		upbitws.SubscriptionMessageFormat{
			Format: "DEFAULT",
		},
	}

	// Send subscription request
	if err := conn.WriteJSON(subscriptionMessage); err != nil {
		return fmt.Errorf("failed to send subscription message: %v", err)
	}

	// Create ping ticker
	pingTicker := time.NewTicker(40 * time.Second)
	defer pingTicker.Stop()

	// Start ping loop in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				// Send ping frame
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("Failed to send ping: %v", err)
					// Closing conn will make ReadJSON fail → triggers reconnect in main loop
					_ = conn.Close()
					return
				} else {
					log.Printf("Ping sent")
				}
			}
		}
	}()

	// Read messages
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var trade upbitws.SpotTrade
			if err := conn.ReadJSON(&trade); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket closed unexpectedly: %v", err)
					// Try to reconnect
					return SubscribeTrade(ctx, symbol, handlers)
				}
				log.Printf("Error reading message: %v", err)
				continue
			}
			for _, handler := range handlers {
				if err := handler(trade); err != nil {
					return fmt.Errorf("handler error: %v", err)
				}
			}
		}
	}
}

func HandlerTrade[T any, V any](
	ch chan T,
	extractor func(upbitws.SpotTrade) V,
	converter func(V) T,
) func(upbitws.SpotTrade) error {
	return func(trade upbitws.SpotTrade) error {
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

func ExtractPrice(trade upbitws.SpotTrade) float64 {
	return trade.TradePrice
}

func ExtractQuantity(trade upbitws.SpotTrade) float64 {
	return trade.TradeVolume
}

func NewPriceHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractPrice, nil)
}

func NewQuantityHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractQuantity, nil)
}
