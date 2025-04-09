package upbitws

// OrderbookUnit represents a single level in the orderbook
type OrderbookUnit struct {
	AskPrice float64 `json:"ask_price"` // Ask price
	BidPrice float64 `json:"bid_price"` // Bid price
	AskSize  float64 `json:"ask_size"`  // Ask size
	BidSize  float64 `json:"bid_size"`  // Bid size
}

// SpotOrderbook represents the spot orderbook data from Upbit websocket
type SpotOrderbook struct {
	Type           string          `json:"type"`            // Type of event
	Code           string          `json:"code"`            // Trading pair code
	Timestamp      int64           `json:"timestamp"`       // Event timestamp
	TotalAskSize   float64         `json:"total_ask_size"`  // Total ask size
	TotalBidSize   float64         `json:"total_bid_size"`  // Total bid size
	OrderbookUnits []OrderbookUnit `json:"orderbook_units"` // Orderbook levels
	StreamType     string          `json:"stream_type"`     // Stream type
	Level          int             `json:"level"`           // Orderbook level
}
