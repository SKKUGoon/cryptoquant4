package binancews

type KlineDataStream struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime    int64  `json:"t"`
		CloseTime    int64  `json:"T"`
		Symbol       string `json:"s"`
		Interval     string `json:"i"`
		FirstTradeID int64  `json:"f"`
		LastTradeID  int64  `json:"L"`
		OpenPrice    string `json:"o"`
		ClosePrice   string `json:"c"`
		HighPrice    string `json:"h"`
		LowPrice     string `json:"l"`
		Volume       string `json:"v"`
		NumTrades    int    `json:"n"`
		IsClosed     bool   `json:"x"`
		QuoteVolume  string `json:"q"`
	} `json:"k"`
}
