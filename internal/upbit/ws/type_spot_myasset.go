package upbitws

// MyAssetResponse represents the response from Upbit myAsset websocket
type MyAssetResponse struct {
	Type           string  `json:"type"`
	AssetUUID      string  `json:"asset_uuid"`
	Assets         []Asset `json:"assets"`
	AssetTimestamp int64   `json:"asset_timestamp"`
	Timestamp      int64   `json:"timestamp"`
	StreamType     string  `json:"stream_type"`
}

// Asset represents an individual asset in the myAsset response
type Asset struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"` // Using string to maintain precision
	Locked   float64 `json:"locked"`  // Using string to maintain precision
}
