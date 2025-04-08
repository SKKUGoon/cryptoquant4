package binancews

// FutureAggTrade represents the aggregate trade data from Binance websocket
type FutureAggTrade struct {
	EventType     string `json:"e"` // Event type
	EventTime     int64  `json:"E"` // Event time
	Symbol        string `json:"s"` // Symbol
	AggTradeID    int64  `json:"a"` // Aggregate trade ID
	Price         string `json:"p"` // Price
	Quantity      string `json:"q"` // Quantity
	FirstTradeID  int64  `json:"f"` // First trade ID
	LastTradeID   int64  `json:"l"` // Last trade ID
	TradeTime     int64  `json:"T"` // Trade time
	IsMarketMaker bool   `json:"m"` // Is the buyer the market maker?
}
