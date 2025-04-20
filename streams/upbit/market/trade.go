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
	maxInitialDelay = 10 * time.Second
)

func SubscribeTrade(ctx context.Context, symbol string, handlers []func(upbitws.SpotTrade) error) error {
	const url = "wss://api.upbit.com/websocket/v1"
	var conn *websocket.Conn
	var err error
	var trade upbitws.SpotTrade
	var backoff = initialBackoff

	// Add random delay before initial connection attempt
	initialDelay := time.Duration(rand.Int63n(int64(maxInitialDelay)))
	log.Printf("Waiting %v before initial connection attempt for %s", initialDelay, symbol)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(initialDelay):
		// Continue with connection attempt
	}

	// Retry connection with exponential backoff
	for retry := range maxRetries {
		if conn, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
			break
		}

		log.Printf("WebSocket connection attempt %d/%d failed: %v", retry+1, maxRetries, err)
		backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))

		// Wait for backoff period or context cancellation
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

	// Upbit requires a ping from the client every 2 minutes. - 40 seconds to be safe
	pingTicker := time.NewTicker(40 * time.Second)
	defer pingTicker.Stop()
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
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

// Best ask and bid price and quantity at execution period.
func ExtractExecBestBidPrc(trade upbitws.SpotTrade) float64 {
	return trade.BestBidPrice
}

func ExtractExecBestBidQty(trade upbitws.SpotTrade) float64 {
	return trade.BestBidSize
}

func ExtractExecBestAskPrc(trade upbitws.SpotTrade) float64 {
	return trade.BestAskPrice
}

func ExtractExecBestAskQty(trade upbitws.SpotTrade) float64 {
	return trade.BestAskSize
}

func NewPriceHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractPrice, nil)
}

func NewQuantityHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractQuantity, nil)
}

func NewExecBestBidPrcHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractExecBestBidPrc, nil)
}

func NewExecBestBidQtyHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractExecBestBidQty, nil)
}

func NewExecBestAskPrcHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractExecBestAskPrc, nil)
}

func NewExecBestAskQtyHandler(ch chan float64) func(upbitws.SpotTrade) error {
	return HandlerTrade(ch, ExtractExecBestAskQty, nil)
}
