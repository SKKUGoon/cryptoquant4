package upbituser

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	upbitws "cryptoquant.com/m/internal/upbit/ws"
	"github.com/gorilla/websocket"
)

const (
	maxRetries     = 5
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

func SubscribeMyAsset(ctx context.Context, handlers []func(upbitws.MyAssetResponse) error) error {
	url := "wss://api.upbit.com/websocket/v1/private"

	var conn *websocket.Conn
	var err error
	backoff := initialBackoff

	// Create private websocket connection with listen key
	_, authToken, err := CreateListenKey()
	if err != nil {
		return fmt.Errorf("failed to create listen key: %v", err)
	}
	header := http.Header{}
	header.Add("Authorization", authToken)

	for retry := range maxRetries {
		if conn, _, err = websocket.DefaultDialer.Dial(url, header); err == nil {
			break // break if connection is successful
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

	log.Printf("Connected to Upbit my asset stream")

	subscriptionMessage := upbitws.SubscriptionMessage{
		upbitws.SubscriptionMessageTicket{
			Ticket: "my_asset_subscription_" + time.Now().Format("20060102"),
		},
		upbitws.SubscriptionMessageType{
			Type: "myAsset",
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

	var myAsset upbitws.MyAssetResponse
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := conn.ReadJSON(&myAsset); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket closed unexpectedly: %v", err)
					return SubscribeMyAsset(ctx, handlers) // retry connection
				}
				log.Printf("Error reading message: %v", err)
				continue
			}
			for _, handler := range handlers {
				if err := handler(myAsset); err != nil {
					return fmt.Errorf("handler error: %v", err)
				}
			}
		}
	}
}

func HandlerMyAsset[T any, V any](
	ch chan T,
	extractor func(upbitws.MyAssetResponse) V,
	converter func(V) T,
) func(upbitws.MyAssetResponse) error {
	return func(myAsset upbitws.MyAssetResponse) error {
		value := extractor(myAsset)

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

func ExtractAssetBalanceKRW(myAsset upbitws.MyAssetResponse) float64 {
	for _, asset := range myAsset.Assets {
		if asset.Currency == "KRW" {
			return asset.Balance
		}
	}

	return 0
}

func NewAssetBalanceKRWHandler(ch chan float64) func(upbitws.MyAssetResponse) error {
	return HandlerMyAsset(ch, ExtractAssetBalanceKRW, nil)
}
