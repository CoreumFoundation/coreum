// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/dex/v1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState defines the module genesis state.
type GenesisState struct {
	// params defines all the parameters of the module.
	Params                     Params                    `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	OrderBooks                 []OrderBookDataWithID     `protobuf:"bytes,2,rep,name=order_books,json=orderBooks,proto3" json:"order_books"`
	Orders                     []OrderWithSequence       `protobuf:"bytes,3,rep,name=orders,proto3" json:"orders"`
	AccountsDenomsOrdersCounts []AccountDenomOrdersCount `protobuf:"bytes,4,rep,name=accounts_denoms_orders_counts,json=accountsDenomsOrdersCounts,proto3" json:"accounts_denoms_orders_counts"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9d24a0566883c25, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetOrderBooks() []OrderBookDataWithID {
	if m != nil {
		return m.OrderBooks
	}
	return nil
}

func (m *GenesisState) GetOrders() []OrderWithSequence {
	if m != nil {
		return m.Orders
	}
	return nil
}

func (m *GenesisState) GetAccountsDenomsOrdersCounts() []AccountDenomOrdersCount {
	if m != nil {
		return m.AccountsDenomsOrdersCounts
	}
	return nil
}

// OrderBookDataWithID is a order book data with it's corresponding ID.
type OrderBookDataWithID struct {
	// id is order book ID.
	ID uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// data is order book data.
	Data OrderBookData `protobuf:"bytes,2,opt,name=data,proto3" json:"data"`
}

func (m *OrderBookDataWithID) Reset()         { *m = OrderBookDataWithID{} }
func (m *OrderBookDataWithID) String() string { return proto.CompactTextString(m) }
func (*OrderBookDataWithID) ProtoMessage()    {}
func (*OrderBookDataWithID) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9d24a0566883c25, []int{1}
}
func (m *OrderBookDataWithID) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OrderBookDataWithID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OrderBookDataWithID.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OrderBookDataWithID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OrderBookDataWithID.Merge(m, src)
}
func (m *OrderBookDataWithID) XXX_Size() int {
	return m.Size()
}
func (m *OrderBookDataWithID) XXX_DiscardUnknown() {
	xxx_messageInfo_OrderBookDataWithID.DiscardUnknown(m)
}

var xxx_messageInfo_OrderBookDataWithID proto.InternalMessageInfo

func (m *OrderBookDataWithID) GetID() uint32 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *OrderBookDataWithID) GetData() OrderBookData {
	if m != nil {
		return m.Data
	}
	return OrderBookData{}
}

// OrderWithSequence is a order with it's corresponding sequence.
type OrderWithSequence struct {
	// sequence is order sequence.
	Sequence uint64 `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	// data is order book data.
	Order Order `protobuf:"bytes,2,opt,name=order,proto3" json:"order"`
}

func (m *OrderWithSequence) Reset()         { *m = OrderWithSequence{} }
func (m *OrderWithSequence) String() string { return proto.CompactTextString(m) }
func (*OrderWithSequence) ProtoMessage()    {}
func (*OrderWithSequence) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9d24a0566883c25, []int{2}
}
func (m *OrderWithSequence) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OrderWithSequence) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OrderWithSequence.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OrderWithSequence) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OrderWithSequence.Merge(m, src)
}
func (m *OrderWithSequence) XXX_Size() int {
	return m.Size()
}
func (m *OrderWithSequence) XXX_DiscardUnknown() {
	xxx_messageInfo_OrderWithSequence.DiscardUnknown(m)
}

var xxx_messageInfo_OrderWithSequence proto.InternalMessageInfo

func (m *OrderWithSequence) GetSequence() uint64 {
	if m != nil {
		return m.Sequence
	}
	return 0
}

func (m *OrderWithSequence) GetOrder() Order {
	if m != nil {
		return m.Order
	}
	return Order{}
}

// AccountDenomOrderCount is a count of orders per account and denom.
type AccountDenomOrdersCount struct {
	AccountNumber uint64 `protobuf:"varint,1,opt,name=account_number,json=accountNumber,proto3" json:"account_number,omitempty"`
	Denom         string `protobuf:"bytes,2,opt,name=denom,proto3" json:"denom,omitempty"`
	OrdersCount   uint64 `protobuf:"varint,3,opt,name=orders_count,json=ordersCount,proto3" json:"orders_count,omitempty"`
}

func (m *AccountDenomOrdersCount) Reset()         { *m = AccountDenomOrdersCount{} }
func (m *AccountDenomOrdersCount) String() string { return proto.CompactTextString(m) }
func (*AccountDenomOrdersCount) ProtoMessage()    {}
func (*AccountDenomOrdersCount) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9d24a0566883c25, []int{3}
}
func (m *AccountDenomOrdersCount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AccountDenomOrdersCount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AccountDenomOrdersCount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AccountDenomOrdersCount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountDenomOrdersCount.Merge(m, src)
}
func (m *AccountDenomOrdersCount) XXX_Size() int {
	return m.Size()
}
func (m *AccountDenomOrdersCount) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountDenomOrdersCount.DiscardUnknown(m)
}

var xxx_messageInfo_AccountDenomOrdersCount proto.InternalMessageInfo

func (m *AccountDenomOrdersCount) GetAccountNumber() uint64 {
	if m != nil {
		return m.AccountNumber
	}
	return 0
}

func (m *AccountDenomOrdersCount) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *AccountDenomOrdersCount) GetOrdersCount() uint64 {
	if m != nil {
		return m.OrdersCount
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "coreum.dex.v1.GenesisState")
	proto.RegisterType((*OrderBookDataWithID)(nil), "coreum.dex.v1.OrderBookDataWithID")
	proto.RegisterType((*OrderWithSequence)(nil), "coreum.dex.v1.OrderWithSequence")
	proto.RegisterType((*AccountDenomOrdersCount)(nil), "coreum.dex.v1.AccountDenomOrdersCount")
}

func init() { proto.RegisterFile("coreum/dex/v1/genesis.proto", fileDescriptor_a9d24a0566883c25) }

var fileDescriptor_a9d24a0566883c25 = []byte{
	// 461 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x8e, 0x9d, 0x34, 0x82, 0x49, 0x83, 0xc4, 0x12, 0xc0, 0x18, 0x70, 0x83, 0x25, 0x50, 0x4f,
	0x36, 0x69, 0x05, 0x47, 0x24, 0xd2, 0x08, 0x14, 0x21, 0x01, 0x4a, 0x0f, 0x48, 0x5c, 0xac, 0x8d,
	0xbd, 0x4a, 0xad, 0xca, 0x9e, 0xe0, 0x5d, 0x47, 0xe9, 0x5b, 0xf0, 0x46, 0x5c, 0x7b, 0xec, 0x91,
	0x53, 0x85, 0x9c, 0x17, 0x41, 0x9e, 0xdd, 0xa0, 0x26, 0x04, 0x6e, 0x3b, 0x33, 0xdf, 0x8f, 0xe7,
	0xf3, 0xc0, 0xe3, 0x18, 0x0b, 0x51, 0x66, 0x61, 0x22, 0x96, 0xe1, 0x62, 0x10, 0xce, 0x44, 0x2e,
	0x64, 0x2a, 0x83, 0x79, 0x81, 0x0a, 0x59, 0x57, 0x0f, 0x83, 0x44, 0x2c, 0x83, 0xc5, 0xc0, 0x7d,
	0xb4, 0x89, 0xc5, 0x22, 0x11, 0x85, 0x46, 0xba, 0xee, 0xe6, 0x68, 0xce, 0x0b, 0x9e, 0x19, 0x15,
	0xb7, 0x37, 0xc3, 0x19, 0xd2, 0x33, 0xac, 0x5f, 0xba, 0xeb, 0xff, 0xb0, 0x61, 0xff, 0xbd, 0x76,
	0x3b, 0x55, 0x5c, 0x09, 0x76, 0x0c, 0x6d, 0x4d, 0x73, 0xac, 0xbe, 0x75, 0xd8, 0x39, 0xba, 0x1f,
	0x6c, 0xb8, 0x07, 0x9f, 0x69, 0x38, 0x6c, 0x5d, 0x5e, 0x1f, 0x34, 0x26, 0x06, 0xca, 0xc6, 0xd0,
	0xa1, 0xcf, 0x88, 0xa6, 0x88, 0xe7, 0xd2, 0xb1, 0xfb, 0xcd, 0xc3, 0xce, 0x91, 0xbf, 0xc5, 0xfc,
	0x54, 0x23, 0x86, 0x88, 0xe7, 0x23, 0xae, 0xf8, 0x97, 0x54, 0x9d, 0x8d, 0x47, 0x46, 0x06, 0x70,
	0x3d, 0x92, 0xec, 0x0d, 0xb4, 0xa9, 0x92, 0x4e, 0x93, 0x54, 0xfa, 0xbb, 0x54, 0x6a, 0xf6, 0xa9,
	0xf8, 0x56, 0x8a, 0x3c, 0x16, 0xeb, 0x4f, 0xd1, 0x2c, 0x86, 0xf0, 0x94, 0xc7, 0x31, 0x96, 0xb9,
	0x92, 0x51, 0x22, 0x72, 0xcc, 0x64, 0xa4, 0x27, 0x91, 0x6e, 0x3a, 0x2d, 0x92, 0x7d, 0xb1, 0x25,
	0xfb, 0x56, 0x73, 0x46, 0x35, 0x83, 0x2c, 0xe4, 0x49, 0x5d, 0x1b, 0x71, 0x77, 0x2d, 0x49, 0x73,
	0x79, 0x03, 0x20, 0x7d, 0x01, 0xf7, 0x76, 0x6c, 0xc6, 0x1e, 0x80, 0x9d, 0x26, 0x94, 0x61, 0x77,
	0xd8, 0xae, 0xae, 0x0f, 0xec, 0xf1, 0x68, 0x62, 0xa7, 0x09, 0x7b, 0x0d, 0xad, 0x84, 0x2b, 0xee,
	0xd8, 0x94, 0xee, 0x93, 0xff, 0x65, 0x64, 0xcc, 0x09, 0xef, 0x73, 0xb8, 0xfb, 0xd7, 0xea, 0xcc,
	0x85, 0x5b, 0xd2, 0xbc, 0xc9, 0xaa, 0x35, 0xf9, 0x53, 0xb3, 0x97, 0xb0, 0x47, 0x8b, 0x1b, 0xa7,
	0xde, 0x4e, 0x27, 0xed, 0xa0, 0x81, 0xfe, 0x05, 0x3c, 0xfc, 0x47, 0x0c, 0xec, 0x39, 0xdc, 0x31,
	0x11, 0x44, 0x79, 0x99, 0x4d, 0x45, 0x61, 0xec, 0xba, 0xa6, 0xfb, 0x91, 0x9a, 0xac, 0x07, 0x7b,
	0x94, 0x39, 0x79, 0xde, 0x9e, 0xe8, 0x82, 0x3d, 0x83, 0xfd, 0x9b, 0xbf, 0xc0, 0x69, 0x12, 0x55,
	0x5f, 0x8c, 0x89, 0xf9, 0xc3, 0x65, 0xe5, 0x59, 0x57, 0x95, 0x67, 0xfd, 0xaa, 0x3c, 0xeb, 0xfb,
	0xca, 0x6b, 0x5c, 0xad, 0xbc, 0xc6, 0xcf, 0x95, 0xd7, 0xf8, 0x3a, 0x98, 0xa5, 0xea, 0xac, 0x9c,
	0x06, 0x31, 0x66, 0xe1, 0x09, 0x6d, 0xf0, 0x0e, 0xcb, 0x3c, 0xe1, 0x2a, 0xc5, 0x3c, 0x34, 0xe7,
	0xbe, 0x78, 0x15, 0x2e, 0xe9, 0xe6, 0xd5, 0xc5, 0x5c, 0xc8, 0x69, 0x9b, 0x4e, 0xfb, 0xf8, 0x77,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x9b, 0xb9, 0x67, 0xcf, 0x55, 0x03, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AccountsDenomsOrdersCounts) > 0 {
		for iNdEx := len(m.AccountsDenomsOrdersCounts) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.AccountsDenomsOrdersCounts[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Orders) > 0 {
		for iNdEx := len(m.Orders) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Orders[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.OrderBooks) > 0 {
		for iNdEx := len(m.OrderBooks) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.OrderBooks[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *OrderBookDataWithID) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OrderBookDataWithID) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OrderBookDataWithID) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Data.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if m.ID != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ID))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *OrderWithSequence) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OrderWithSequence) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OrderWithSequence) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Order.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if m.Sequence != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Sequence))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *AccountDenomOrdersCount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AccountDenomOrdersCount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AccountDenomOrdersCount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.OrdersCount != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.OrdersCount))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0x12
	}
	if m.AccountNumber != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.AccountNumber))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.OrderBooks) > 0 {
		for _, e := range m.OrderBooks {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Orders) > 0 {
		for _, e := range m.Orders {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.AccountsDenomsOrdersCounts) > 0 {
		for _, e := range m.AccountsDenomsOrdersCounts {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *OrderBookDataWithID) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ID != 0 {
		n += 1 + sovGenesis(uint64(m.ID))
	}
	l = m.Data.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *OrderWithSequence) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Sequence != 0 {
		n += 1 + sovGenesis(uint64(m.Sequence))
	}
	l = m.Order.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *AccountDenomOrdersCount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.AccountNumber != 0 {
		n += 1 + sovGenesis(uint64(m.AccountNumber))
	}
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.OrdersCount != 0 {
		n += 1 + sovGenesis(uint64(m.OrdersCount))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrderBooks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrderBooks = append(m.OrderBooks, OrderBookDataWithID{})
			if err := m.OrderBooks[len(m.OrderBooks)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Orders", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Orders = append(m.Orders, OrderWithSequence{})
			if err := m.Orders[len(m.Orders)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountsDenomsOrdersCounts", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountsDenomsOrdersCounts = append(m.AccountsDenomsOrdersCounts, AccountDenomOrdersCount{})
			if err := m.AccountsDenomsOrdersCounts[len(m.AccountsDenomsOrdersCounts)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *OrderBookDataWithID) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: OrderBookDataWithID: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OrderBookDataWithID: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Data.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *OrderWithSequence) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: OrderWithSequence: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OrderWithSequence: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sequence", wireType)
			}
			m.Sequence = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Sequence |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Order", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Order.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *AccountDenomOrdersCount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AccountDenomOrdersCount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AccountDenomOrdersCount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountNumber", wireType)
			}
			m.AccountNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AccountNumber |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrdersCount", wireType)
			}
			m.OrdersCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.OrdersCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
