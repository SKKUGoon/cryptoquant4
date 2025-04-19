package binancerest

type LeverageRequest struct {
	Symbol    string `json:"symbol"`
	Leverage  int    `json:"leverage"`
	Timestamp int64  `json:"timestamp"`
}

type LeverageResponse struct {
	Leverage         int    `json:"leverage"`
	MaxNotionalValue string `json:"maxNotionalValue"`
	Symbol           string `json:"symbol"`
}
