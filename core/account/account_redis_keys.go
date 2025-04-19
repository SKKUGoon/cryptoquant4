package account

import "fmt"

// ReservedFund: Total amount of money (USDT/KRW) that is reserved by containers.
func (a *AccountSource) keyReservedFund(exchange string) string {
	return fmt.Sprintf("reserved_fund:%s", exchange)
}

// AvailableFund: Total amount of money (USDT/KRW) that is free to be reserved by containers.
func (a *AccountSource) keyAvailableFund(exchange string) string {
	return fmt.Sprintf("available_fund:%s", exchange)
}

// Position: Total amount of money (USDT/KRW) that is reserved by containers.
func (a *AccountSource) keyPosition(exchange, currency string) string {
	return fmt.Sprintf("wallet:%s:%s", exchange, currency)
}

// WalletSnapshot: Snapshot of the wallet at startup.
func (a *AccountSource) keyWalletSnapshot(exchange string) string {
	return fmt.Sprintf("wallet_snapshot:%s", exchange)
}

// OrderId: Order id
func (a *AccountSource) keyOrderId(exchange string) string {
	return fmt.Sprintf("order_id:%s", exchange)
}
