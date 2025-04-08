package binancews

// FutureBookTicker represents the book ticker data from Binance websocket stream
type FutureBookTicker struct {
	EventType       string `json:"e"` // Event type
	UpdateID        int64  `json:"u"` // Order book updateId
	EventTime       int64  `json:"E"` // Event time
	TransactionTime int64  `json:"T"` // Transaction time
	Symbol          string `json:"s"` // Symbol
	BestBidPrice    string `json:"b"` // Best bid price
	BestBidQty      string `json:"B"` // Best bid quantity
	BestAskPrice    string `json:"a"` // Best ask price
	BestAskQty      string `json:"A"` // Best ask quantity
}
