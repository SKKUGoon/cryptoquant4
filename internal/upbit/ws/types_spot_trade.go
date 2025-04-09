package upbitws

type SubscriptionMessage []any

type SubscriptionMessageTicket struct {
	Ticket string `json:"ticket,omitempty"`
}

type SubscriptionMessageType struct {
	Type  string   `json:"type,omitempty"`
	Codes []string `json:"codes,omitempty"`
	Level int      `json:"level,omitempty"`
}

type SubscriptionMessageFormat struct {
	Format string `json:"format,omitempty"`
}

// SpotTrade represents the spot trade data from Upbit websocket
type SpotTrade struct {
	Type             string  `json:"type"`               // Type of event
	Code             string  `json:"code"`               // Trading pair code
	Timestamp        int64   `json:"timestamp"`          // Event timestamp
	TradeDate        string  `json:"trade_date"`         // Trade date
	TradeTime        string  `json:"trade_time"`         // Trade time
	TradeTimestamp   int64   `json:"trade_timestamp"`    // Trade timestamp
	TradePrice       float64 `json:"trade_price"`        // Trade price
	TradeVolume      float64 `json:"trade_volume"`       // Trade volume
	AskBid           string  `json:"ask_bid"`            // Ask or Bid
	PrevClosingPrice float64 `json:"prev_closing_price"` // Previous closing price
	Change           string  `json:"change"`             // Price change direction
	ChangePrice      float64 `json:"change_price"`       // Price change amount
	SequentialID     int64   `json:"sequential_id"`      // Sequential ID
	BestAskPrice     float64 `json:"best_ask_price"`     // Best ask price
	BestAskSize      float64 `json:"best_ask_size"`      // Best ask size
	BestBidPrice     float64 `json:"best_bid_price"`     // Best bid price
	BestBidSize      float64 `json:"best_bid_size"`      // Best bid size
	StreamType       string  `json:"stream_type"`        // Stream type
}
