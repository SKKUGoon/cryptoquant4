package binanceuser

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"

	binancews "cryptoquant.com/m/internal/binance/ws"
)

// SubscribeBalance subscribes to the balance stream for a given listen key
// and sends the data to the channel
func SubscribeBalance(listenKey string, ch chan binancews.AccountUpdateEvent, done chan struct{}) error {
	const urlBase = "wss://fstream.binance.com/ws/"
	var streamData binancews.AccountUpdateEvent

	url := urlBase + listenKey
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to Binance Futures user stream for %s", listenKey)

	for {
		select {
		case <-done:
			err := conn.WriteMessage(websocket.CloseMessage, []byte{})
			if err != nil {
				log.Printf("Error closing connection: %v", err)
			}
			println("Connection to Binance Futures user stream for %s is closed", listenKey)
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("error reading message: %v", err)
			}

			if err := json.Unmarshal(message, &streamData); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}
			log.Printf("[balance] %v", streamData)
			ch <- streamData
		}
	}
}
