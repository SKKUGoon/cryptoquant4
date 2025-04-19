package signal

import (
	"fmt"
	"log"
	"os"
)

// ConfirmTargetSymbols verifies that the trading symbols specified in environment variables
// are available for trading on both Binance and Upbit exchanges. It checks:
// 1. BINANCE_SYMBOL and UPBIT_SYMBOL environment variables are set
// 2. The symbols are available for trading on their respective exchanges
// 3. The anchor asset symbol is available on Upbit
// If any check fails, it logs an error and panics. Otherwise, it sets the confirmed
// symbols in the engine context.
func (e *SignalContext) ConfirmTargetSymbols() {
	csymbol := os.Getenv("BINANCE_SYMBOL")
	ksymbol := os.Getenv("UPBIT_SYMBOL")

	if csymbol == "" || ksymbol == "" {
		log.Println("Failed to confirm target symbols: Environment variables not set")
		panic("Environment variables not set")
	}

	// Confirm trading symbols in cefi and kimchi
	if !e.BinanceExchangeConfig.IsAvailableSymbol(csymbol) {
		log.Println("Failed to confirm target symbols: Binance symbol not available")
		panic("Binance symbol not available")
	}

	if !e.UpbitExchangeConfig.IsAvailableSymbol(ksymbol) {
		log.Println("Failed to confirm target symbols: Kimchi symbol not available")
		panic("Kimchi symbol not available")
	}

	// Confirm anchor symbol
	if !e.UpbitExchangeConfig.IsAvailableSymbol(e.AnchorAssetSymbol) {
		log.Println("Failed to confirm target symbols: Kimchi anchor symbol not available")
		panic("Kimchi anchor symbol not available")
	}

	e.UpbitAssetSymbol = ksymbol
	e.BinanceAssetSymbol = csymbol
}

// ConfirmTradeParameters retrieves trade parameters from the database and sets them in the engine context.
// It retrieves: (as a default)
// 1. EnterPremiumBoundary: 0.9980
// 2. ExitPremiumBoundary: 1.0035
// These parameters are used to determine the entry and exit points for the arbitrage strategy.
func (e *SignalContext) ConfirmTradeParameters() {
	log.Println("Confirming trade parameters")

	// Premium Calculation Parameters key value
	enterPremiumBoundaryKey := fmt.Sprintf("%v_enter_premium_boundary", e.SignalID)
	exitPremiumBoundaryKey := fmt.Sprintf("%v_exit_premium_boundary", e.SignalID)

	// Premium Calculation Parameters

	// EnterPremiumBoundary: 0.9980
	enterPremiumBoundary, err := e.Database.GetTradeMetadata(enterPremiumBoundaryKey, 0.9980)
	if err != nil {
		log.Printf("Failed to get enter premium boundary: %v", err)
		panic(err)
	}
	e.EnterPremiumBoundary = enterPremiumBoundary.(float64)

	// ExitPremiumBoundary: 1.0035
	exitPremiumBoundary, err := e.Database.GetTradeMetadata(exitPremiumBoundaryKey, 1.0035)
	if err != nil {
		log.Printf("Failed to get exit premium boundary: %v", err)
		panic(err)
	}
	e.ExitPremiumBoundary = exitPremiumBoundary.(float64)

	log.Printf("Trade parameters confirmed: enterPremiumBoundary: %v, exitPremiumBoundary: %v", e.EnterPremiumBoundary, e.ExitPremiumBoundary)
}
