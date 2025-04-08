package upbitmarket

import (
	"context"
	"fmt"
	"log"
	"time"

	upbitws "cryptoquant.com/m/internal/upbit/ws"
	"github.com/gorilla/websocket"
)

func SubscribeTrade(ctx context.Context, symbol string, handlers []func(upbitws.SpotTrade) error) error {
	// Upbit Websocket endpoint
	url := "wss://api.upbit.com/websocket/v1"

	// Connect to Websocket
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
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

	// Read messages
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pingTicker.C:
			// Send ping frame
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				return fmt.Errorf("failed to send ping: %v", err)
			}
		default:
			var trade upbitws.SpotTrade
			if err := conn.ReadJSON(&trade); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					return fmt.Errorf("websocket closed unexpectedly: %v", err)
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
