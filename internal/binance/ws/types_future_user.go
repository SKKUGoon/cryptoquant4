package binancews

type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

type AccountUpdateEvent struct {
	EventType         string            `json:"e"` // ACCOUNT_UPDATE for account update, ORDER_TRADE_UPDATE for order trade update
	EventTime         int64             `json:"E"`
	TransactionTime   int64             `json:"T"`
	AccountUpdateData AccountUpdateData `json:"a,omitempty"`
	OrderUpdateData   OrderTradeUpdata  `json:"o,omitempty"`
}

type AccountUpdateData struct {
	// DEPOSIT, WITHDRAW, ORDER, FUNDING_FEE, WITHDRAW_REJECT, ADJUSTMENT, INSURANCE_CLEAR
	// ADMIN_DEPOSIT, ADMIN_WITHDRAW, MARGIN_TRANSFER, MARGIN_TYPE_CHANGE, ASSET_TRANSFER
	// OPTIONS_PREMIUM_FEE, OPTIONS_SETTLE_PROFIT, AUTO_EXCHANGE, COIN_SWAP_DEPOSIT, COIN_SWAP_WITHDRAW
	Reason    string           `json:"m"`
	Balances  []BalanceUpdate  `json:"B"`
	Positions []PositionUpdate `json:"P"`
}

type BalanceUpdate struct {
	Asset              string `json:"a"`
	WalletBalance      string `json:"wb"`
	CrossWalletBalance string `json:"cw"`
	BalanceChange      string `json:"bc"`
}

type PositionUpdate struct {
	Symbol         string `json:"s"`
	PositionAmount string `json:"pa"`
	EntryPrice     string `json:"ep"`
	BreakEvenPrice string `json:"bep"`
	CumRealized    string `json:"cr"`
	UnrealizedPnL  string `json:"up"`
	MarginType     string `json:"mt"`
	IsolatedWallet string `json:"iw"`
	PositionSide   string `json:"ps"`
}

type OrderTradeUpdata struct {
	Symbol              string `json:"s"`
	ClientOrderID       string `json:"c"`
	Side                string `json:"S"` // BUY, SELL
	OrderType           string `json:"o"` // LIMIT, MARKET, STOP, STOP_MARKET, TAKE_PROFIT, TAKE_PROFIT_MARKET, TRAILING_STOP_MARKET, LIQUIDATION
	TimeInForce         string `json:"f"` // GTC, IOC, FOK, GTX
	OriginalQuantity    string `json:"q"`
	OriginalPrice       string `json:"p"`
	AveragePrice        string `json:"ap"`
	StopPrice           string `json:"sp"`
	ExecutionType       string `json:"x"` // NEW, CANCELED, CALCULATED(Liquidation), EXPIRED, TRADE, AMENDMENT(Order modified)
	OrderStatus         string `json:"X"` // NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED, EXPIRED, EXPIRED_IN_MATCH
	OrderID             int64  `json:"i"`
	LastFilledQuantity  string `json:"l"`
	AccumulatedQuantity string `json:"z"`
	LastFilledPrice     string `json:"L"`
	CommissionAsset     string `json:"N,omitempty"`
	Commission          string `json:"n,omitempty"`
	OrderTradeTime      int64  `json:"T"`
	TradeID             int64  `json:"t"`
	BidsNotional        string `json:"b"`
	AskNotional         string `json:"a"`
	IsMaker             bool   `json:"m"`
	IsReduceOnly        bool   `json:"R"`
	WorkingType         string `json:"wt"` // MARK_PRICE, CONTRACT_PRICE
	OriginalOrderType   string `json:"ot"`
	PositionSide        string `json:"ps"`
	IsCloseAll          bool   `json:"cp"`
	ActivationPrice     string `json:"AP,omitempty"`
	CallbackRate        string `json:"cr,omitempty"`
	PriceProtect        bool   `json:"pP"`
	RealizedProfit      string `json:"rp"`
	STPMode             string `json:"V"`
	PriceMatchMode      string `json:"pm"`
	GTDCancelTime       int64  `json:"gtd"`
}
