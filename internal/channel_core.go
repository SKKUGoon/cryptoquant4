package internal

// Global channels for pipeline communication

var (
	// PriceDataChannel   = make(chan PriceData)
	// TradeSignalChannel = make(chan TradeSignal)
	// OrderChannel       = make(chan Order)
	ErrorChan = make(chan error)
)
