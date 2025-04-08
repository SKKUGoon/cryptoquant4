package upbitws

// AssetUpdate represents the user asset update data from Upbit websocket
type AssetUpdate struct {
	Type           string        `json:"type"`            // Type of event
	AssetUUID      string        `json:"asset_uuid"`      // Asset UUID
	Assets         []AssetDetail `json:"assets"`          // List of assets
	AssetTimestamp int64         `json:"asset_timestamp"` // Asset timestamp
	Timestamp      int64         `json:"timestamp"`       // Event timestamp
	StreamType     string        `json:"stream_type"`     // Stream type
}

// AssetDetail represents a single asset detail in an asset update
type AssetDetail struct {
	Currency string  `json:"currency"` // Currency code
	Balance  float64 `json:"balance"`  // Available balance
	Locked   float64 `json:"locked"`   // Locked balance (in orders)
}
