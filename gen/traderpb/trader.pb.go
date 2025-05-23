// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.28.3
// source: trader.proto

package traderpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// --- ENUM for clarity ---
type PairOrderType int32

const (
	PairOrderType_PairOrderTypeUnspecified PairOrderType = 0
	PairOrderType_PairOrderEnter           PairOrderType = 1 // Bet on Long Premium
	PairOrderType_PairOrderExit            PairOrderType = 2 // Bet on Short Premium
)

// Enum value maps for PairOrderType.
var (
	PairOrderType_name = map[int32]string{
		0: "PairOrderTypeUnspecified",
		1: "PairOrderEnter",
		2: "PairOrderExit",
	}
	PairOrderType_value = map[string]int32{
		"PairOrderTypeUnspecified": 0,
		"PairOrderEnter":           1,
		"PairOrderExit":            2,
	}
)

func (x PairOrderType) Enum() *PairOrderType {
	p := new(PairOrderType)
	*p = x
	return p
}

func (x PairOrderType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PairOrderType) Descriptor() protoreflect.EnumDescriptor {
	return file_trader_proto_enumTypes[0].Descriptor()
}

func (PairOrderType) Type() protoreflect.EnumType {
	return &file_trader_proto_enumTypes[0]
}

func (x PairOrderType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PairOrderType.Descriptor instead.
func (PairOrderType) EnumDescriptor() ([]byte, []int) {
	return file_trader_proto_rawDescGZIP(), []int{0}
}

// --- Generalized exchange order ---
type ExchangeOrder struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exchange string  `protobuf:"bytes,1,opt,name=exchange,proto3" json:"exchange,omitempty"` // "upbit", "binance", etc.
	Symbol   string  `protobuf:"bytes,2,opt,name=symbol,proto3" json:"symbol,omitempty"`     // e.g., "BTCUSDT" for binance, "BTC-KRW" for upbit
	Side     string  `protobuf:"bytes,3,opt,name=side,proto3" json:"side,omitempty"`         // "buy" or "sell"
	Price    float64 `protobuf:"fixed64,4,opt,name=price,proto3" json:"price,omitempty"`     // Best Bid or Ask price
	Amount   float64 `protobuf:"fixed64,5,opt,name=amount,proto3" json:"amount,omitempty"`   // Best Bid or Ask amount
}

func (x *ExchangeOrder) Reset() {
	*x = ExchangeOrder{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trader_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExchangeOrder) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExchangeOrder) ProtoMessage() {}

func (x *ExchangeOrder) ProtoReflect() protoreflect.Message {
	mi := &file_trader_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExchangeOrder.ProtoReflect.Descriptor instead.
func (*ExchangeOrder) Descriptor() ([]byte, []int) {
	return file_trader_proto_rawDescGZIP(), []int{0}
}

func (x *ExchangeOrder) GetExchange() string {
	if x != nil {
		return x.Exchange
	}
	return ""
}

func (x *ExchangeOrder) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *ExchangeOrder) GetSide() string {
	if x != nil {
		return x.Side
	}
	return ""
}

func (x *ExchangeOrder) GetPrice() float64 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *ExchangeOrder) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

type SingleOrderSheet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Symbol string         `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Order  *ExchangeOrder `protobuf:"bytes,2,opt,name=order,proto3" json:"order,omitempty"`
	Reason string         `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (x *SingleOrderSheet) Reset() {
	*x = SingleOrderSheet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trader_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SingleOrderSheet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SingleOrderSheet) ProtoMessage() {}

func (x *SingleOrderSheet) ProtoReflect() protoreflect.Message {
	mi := &file_trader_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SingleOrderSheet.ProtoReflect.Descriptor instead.
func (*SingleOrderSheet) Descriptor() ([]byte, []int) {
	return file_trader_proto_rawDescGZIP(), []int{1}
}

func (x *SingleOrderSheet) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *SingleOrderSheet) GetOrder() *ExchangeOrder {
	if x != nil {
		return x.Order
	}
	return nil
}

func (x *SingleOrderSheet) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

type PairOrderSheet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BaseSymbol    string         `protobuf:"bytes,1,opt,name=base_symbol,json=baseSymbol,proto3" json:"base_symbol,omitempty"`
	ExchangeRate  float64        `protobuf:"fixed64,2,opt,name=exchange_rate,json=exchangeRate,proto3" json:"exchange_rate,omitempty"`
	PairOrderType PairOrderType  `protobuf:"varint,3,opt,name=pair_order_type,json=pairOrderType,proto3,enum=trader.PairOrderType" json:"pair_order_type,omitempty"`
	UpbitOrder    *ExchangeOrder `protobuf:"bytes,4,opt,name=upbit_order,json=upbitOrder,proto3" json:"upbit_order,omitempty"`
	BinanceOrder  *ExchangeOrder `protobuf:"bytes,5,opt,name=binance_order,json=binanceOrder,proto3" json:"binance_order,omitempty"`
	Reason        string         `protobuf:"bytes,6,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (x *PairOrderSheet) Reset() {
	*x = PairOrderSheet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trader_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PairOrderSheet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairOrderSheet) ProtoMessage() {}

func (x *PairOrderSheet) ProtoReflect() protoreflect.Message {
	mi := &file_trader_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairOrderSheet.ProtoReflect.Descriptor instead.
func (*PairOrderSheet) Descriptor() ([]byte, []int) {
	return file_trader_proto_rawDescGZIP(), []int{2}
}

func (x *PairOrderSheet) GetBaseSymbol() string {
	if x != nil {
		return x.BaseSymbol
	}
	return ""
}

func (x *PairOrderSheet) GetExchangeRate() float64 {
	if x != nil {
		return x.ExchangeRate
	}
	return 0
}

func (x *PairOrderSheet) GetPairOrderType() PairOrderType {
	if x != nil {
		return x.PairOrderType
	}
	return PairOrderType_PairOrderTypeUnspecified
}

func (x *PairOrderSheet) GetUpbitOrder() *ExchangeOrder {
	if x != nil {
		return x.UpbitOrder
	}
	return nil
}

func (x *PairOrderSheet) GetBinanceOrder() *ExchangeOrder {
	if x != nil {
		return x.BinanceOrder
	}
	return nil
}

func (x *PairOrderSheet) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

type TradeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to OrderType:
	//
	//	*TradeRequest_SingleOrder
	//	*TradeRequest_PairOrder
	OrderType isTradeRequest_OrderType `protobuf_oneof:"order_type"`
}

func (x *TradeRequest) Reset() {
	*x = TradeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trader_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TradeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TradeRequest) ProtoMessage() {}

func (x *TradeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_trader_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TradeRequest.ProtoReflect.Descriptor instead.
func (*TradeRequest) Descriptor() ([]byte, []int) {
	return file_trader_proto_rawDescGZIP(), []int{3}
}

func (m *TradeRequest) GetOrderType() isTradeRequest_OrderType {
	if m != nil {
		return m.OrderType
	}
	return nil
}

func (x *TradeRequest) GetSingleOrder() *SingleOrderSheet {
	if x, ok := x.GetOrderType().(*TradeRequest_SingleOrder); ok {
		return x.SingleOrder
	}
	return nil
}

func (x *TradeRequest) GetPairOrder() *PairOrderSheet {
	if x, ok := x.GetOrderType().(*TradeRequest_PairOrder); ok {
		return x.PairOrder
	}
	return nil
}

type isTradeRequest_OrderType interface {
	isTradeRequest_OrderType()
}

type TradeRequest_SingleOrder struct {
	SingleOrder *SingleOrderSheet `protobuf:"bytes,1,opt,name=single_order,json=singleOrder,proto3,oneof"`
}

type TradeRequest_PairOrder struct {
	PairOrder *PairOrderSheet `protobuf:"bytes,2,opt,name=pair_order,json=pairOrder,proto3,oneof"` // Future: MultiLegOrderSheet multi_leg_order = 3;
}

func (*TradeRequest_SingleOrder) isTradeRequest_OrderType() {}

func (*TradeRequest_PairOrder) isTradeRequest_OrderType() {}

type OrderResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *OrderResponse) Reset() {
	*x = OrderResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trader_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OrderResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderResponse) ProtoMessage() {}

func (x *OrderResponse) ProtoReflect() protoreflect.Message {
	mi := &file_trader_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderResponse.ProtoReflect.Descriptor instead.
func (*OrderResponse) Descriptor() ([]byte, []int) {
	return file_trader_proto_rawDescGZIP(), []int{4}
}

func (x *OrderResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *OrderResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_trader_proto protoreflect.FileDescriptor

var file_trader_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x22, 0x85, 0x01, 0x0a, 0x0d, 0x45, 0x78, 0x63, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x78, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x78, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x69, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x73, 0x69, 0x64, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x6f,
	0x0a, 0x10, 0x53, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53, 0x68, 0x65,
	0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x2b, 0x0a, 0x05, 0x6f, 0x72,
	0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x72, 0x61, 0x64,
	0x65, 0x72, 0x2e, 0x45, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x72, 0x64, 0x65, 0x72,
	0x52, 0x05, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x22,
	0xa1, 0x02, 0x0a, 0x0e, 0x50, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53, 0x68, 0x65,
	0x65, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x79, 0x6d, 0x62, 0x6f,
	0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62, 0x61, 0x73, 0x65, 0x53, 0x79, 0x6d,
	0x62, 0x6f, 0x6c, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f,
	0x72, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0c, 0x65, 0x78, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x52, 0x61, 0x74, 0x65, 0x12, 0x3d, 0x0a, 0x0f, 0x70, 0x61, 0x69, 0x72,
	0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x15, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x50, 0x61, 0x69, 0x72, 0x4f,
	0x72, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0d, 0x70, 0x61, 0x69, 0x72, 0x4f, 0x72,
	0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x36, 0x0a, 0x0b, 0x75, 0x70, 0x62, 0x69, 0x74,
	0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74,
	0x72, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x45, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x72,
	0x64, 0x65, 0x72, 0x52, 0x0a, 0x75, 0x70, 0x62, 0x69, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12,
	0x3a, 0x0a, 0x0d, 0x62, 0x69, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x2e,
	0x45, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x52, 0x0c, 0x62,
	0x69, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x22, 0x94, 0x01, 0x0a, 0x0c, 0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x0c, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x5f, 0x6f,
	0x72, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x74, 0x72, 0x61,
	0x64, 0x65, 0x72, 0x2e, 0x53, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53,
	0x68, 0x65, 0x65, 0x74, 0x48, 0x00, 0x52, 0x0b, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x4f, 0x72,
	0x64, 0x65, 0x72, 0x12, 0x37, 0x0a, 0x0a, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x6f, 0x72, 0x64, 0x65,
	0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72,
	0x2e, 0x50, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53, 0x68, 0x65, 0x65, 0x74, 0x48,
	0x00, 0x52, 0x09, 0x70, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x0c, 0x0a, 0x0a,
	0x6f, 0x72, 0x64, 0x65, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x22, 0x43, 0x0a, 0x0d, 0x4f, 0x72,
	0x64, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2a,
	0x54, 0x0a, 0x0d, 0x50, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x1c, 0x0a, 0x18, 0x50, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70,
	0x65, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65, 0x64, 0x10, 0x00, 0x12, 0x12,
	0x0a, 0x0e, 0x50, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x65, 0x72,
	0x10, 0x01, 0x12, 0x11, 0x0a, 0x0d, 0x50, 0x61, 0x69, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x45,
	0x78, 0x69, 0x74, 0x10, 0x02, 0x32, 0x44, 0x0a, 0x06, 0x54, 0x72, 0x61, 0x64, 0x65, 0x72, 0x12,
	0x3a, 0x0a, 0x0b, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x54, 0x72, 0x61, 0x64, 0x65, 0x12, 0x14,
	0x2e, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x4f, 0x72,
	0x64, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x20, 0x5a, 0x1e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x6f, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74, 0x72, 0x61, 0x64, 0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_trader_proto_rawDescOnce sync.Once
	file_trader_proto_rawDescData = file_trader_proto_rawDesc
)

func file_trader_proto_rawDescGZIP() []byte {
	file_trader_proto_rawDescOnce.Do(func() {
		file_trader_proto_rawDescData = protoimpl.X.CompressGZIP(file_trader_proto_rawDescData)
	})
	return file_trader_proto_rawDescData
}

var file_trader_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_trader_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_trader_proto_goTypes = []interface{}{
	(PairOrderType)(0),       // 0: trader.PairOrderType
	(*ExchangeOrder)(nil),    // 1: trader.ExchangeOrder
	(*SingleOrderSheet)(nil), // 2: trader.SingleOrderSheet
	(*PairOrderSheet)(nil),   // 3: trader.PairOrderSheet
	(*TradeRequest)(nil),     // 4: trader.TradeRequest
	(*OrderResponse)(nil),    // 5: trader.OrderResponse
}
var file_trader_proto_depIdxs = []int32{
	1, // 0: trader.SingleOrderSheet.order:type_name -> trader.ExchangeOrder
	0, // 1: trader.PairOrderSheet.pair_order_type:type_name -> trader.PairOrderType
	1, // 2: trader.PairOrderSheet.upbit_order:type_name -> trader.ExchangeOrder
	1, // 3: trader.PairOrderSheet.binance_order:type_name -> trader.ExchangeOrder
	2, // 4: trader.TradeRequest.single_order:type_name -> trader.SingleOrderSheet
	3, // 5: trader.TradeRequest.pair_order:type_name -> trader.PairOrderSheet
	4, // 6: trader.Trader.SubmitTrade:input_type -> trader.TradeRequest
	5, // 7: trader.Trader.SubmitTrade:output_type -> trader.OrderResponse
	7, // [7:8] is the sub-list for method output_type
	6, // [6:7] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_trader_proto_init() }
func file_trader_proto_init() {
	if File_trader_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_trader_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExchangeOrder); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_trader_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SingleOrderSheet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_trader_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PairOrderSheet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_trader_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TradeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_trader_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OrderResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_trader_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*TradeRequest_SingleOrder)(nil),
		(*TradeRequest_PairOrder)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_trader_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_trader_proto_goTypes,
		DependencyIndexes: file_trader_proto_depIdxs,
		EnumInfos:         file_trader_proto_enumTypes,
		MessageInfos:      file_trader_proto_msgTypes,
	}.Build()
	File_trader_proto = out.File
	file_trader_proto_rawDesc = nil
	file_trader_proto_goTypes = nil
	file_trader_proto_depIdxs = nil
}
