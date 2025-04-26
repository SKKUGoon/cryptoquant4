package database

import "time"

type PremiumLog struct {
	Timestamp   time.Time
	Symbol      string
	AnchorPrice float64
	CefiPrice   float64
	// Best Bid and Ask
	KimchiBestBid float64
	KimchiBestAsk float64
	CefiBestBid   float64
	CefiBestAsk   float64
}

type KimchiOrderLog struct {
	PairID        string
	OrderTime     time.Time
	ExecutionTime time.Time
	Symbol        string
	PairSide      string
	Exchange      string
	Side          string
	OrderPrice    float64
	ExecutedPrice float64
	AnchorPrice   float64
}

type AccountSnapshot struct {
	Timestamp         time.Time
	Exchange          string
	Available         float64
	Reserved          float64
	Total             float64
	WalletBalanceUSDT float64
	WalletBalanceKRW  float64
}
