package binancerest

import (
	"time"

	"cryptoquant.com/m/utils"
)

type OrderQueryAll struct {
	Symbol     string `json:"symbol"`
	OrderID    int64  `json:"orderId,omitempty"`
	StartTime  int64  `json:"startTime,omitempty"`
	EndTime    int64  `json:"endTime,omitempty"`
	Limit      int    `json:"limit,omitempty"` // default 500, max 1000
	RecvWindow int64  `json:"recvWindow,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}

func NewOrderQuery() *OrderQueryAll {
	return &OrderQueryAll{
		Timestamp: time.Now().UnixMilli(),
	}
}

func (o *OrderQueryAll) ToParamsMap() map[string]string {
	return utils.StructToParamsMap(o)
}
