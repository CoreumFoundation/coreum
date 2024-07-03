// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/dex/v1/order.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
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

// Side is order side.
type Side int32

const (
	Side_unspecified Side = 0
	Side_sell        Side = 1
	Side_buy         Side = 2
)

var Side_name = map[int32]string{
	0: "unspecified",
	1: "sell",
	2: "buy",
}

var Side_value = map[string]int32{
	"unspecified": 0,
	"sell":        1,
	"buy":         2,
}

func (x Side) String() string {
	return proto.EnumName(Side_name, int32(x))
}

func (Side) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_302bb6c9a553771c, []int{0}
}

// Order is a DEX order.
type Order struct {
	// id is unique order ID.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *Order) Reset()         { *m = Order{} }
func (m *Order) String() string { return proto.CompactTextString(m) }
func (*Order) ProtoMessage()    {}
func (*Order) Descriptor() ([]byte, []int) {
	return fileDescriptor_302bb6c9a553771c, []int{0}
}
func (m *Order) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Order) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Order.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Order) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Order.Merge(m, src)
}
func (m *Order) XXX_Size() int {
	return m.Size()
}
func (m *Order) XXX_DiscardUnknown() {
	xxx_messageInfo_Order.DiscardUnknown(m)
}

var xxx_messageInfo_Order proto.InternalMessageInfo

// OrderBookRecord is a single order book record.
type OrderBookRecord struct {
	// pairID is tokens pair ID.
	PairID uint64 `protobuf:"varint,1,opt,name=pairID,proto3" json:"pairID,omitempty"`
	// side is order side.
	Side Side `protobuf:"varint,2,opt,name=side,proto3,enum=coreum.dex.v1.Side" json:"side,omitempty"`
	// price is order book record price.
	Price Price `protobuf:"bytes,3,opt,name=price,proto3,customtype=Price" json:"price"`
	// order_seq is order sequence.
	OrderSeq uint64 `protobuf:"varint,4,opt,name=order_seq,json=orderSeq,proto3" json:"order_seq,omitempty"`
	// order ID provided by the account.
	OrderID string `protobuf:"bytes,5,opt,name=orderID,proto3" json:"orderID,omitempty"`
	// accountID is account ID which corresponds the order creator.
	AccountID string `protobuf:"bytes,6,opt,name=accountID,proto3" json:"accountID,omitempty"`
	// remaining_quantity is remaining filling quantity sell/buy.
	RemainingQuantity cosmossdk_io_math.Int `protobuf:"bytes,7,opt,name=remaining_quantity,json=remainingQuantity,proto3,customtype=cosmossdk.io/math.Int" json:"remaining_quantity"`
	// remaining_balance is remaining order balance.
	RemainingBalance cosmossdk_io_math.Int `protobuf:"bytes,8,opt,name=remaining_balance,json=remainingBalance,proto3,customtype=cosmossdk.io/math.Int" json:"remaining_balance"`
}

func (m *OrderBookRecord) Reset()         { *m = OrderBookRecord{} }
func (m *OrderBookRecord) String() string { return proto.CompactTextString(m) }
func (*OrderBookRecord) ProtoMessage()    {}
func (*OrderBookRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_302bb6c9a553771c, []int{1}
}
func (m *OrderBookRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OrderBookRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OrderBookRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OrderBookRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OrderBookRecord.Merge(m, src)
}
func (m *OrderBookRecord) XXX_Size() int {
	return m.Size()
}
func (m *OrderBookRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_OrderBookRecord.DiscardUnknown(m)
}

var xxx_messageInfo_OrderBookRecord proto.InternalMessageInfo

// OrderBookStoreRecord is a single order book record used for the store.
type OrderBookStoreRecord struct {
	// order ID provided by the account.
	OrderID string `protobuf:"bytes,4,opt,name=orderID,proto3" json:"orderID,omitempty"`
	// accountID is account ID which corresponds the order creator.
	AccountID string `protobuf:"bytes,5,opt,name=accountID,proto3" json:"accountID,omitempty"`
	// remaining_quantity is remaining filling quantity sell/buy.
	RemainingQuantity cosmossdk_io_math.Int `protobuf:"bytes,6,opt,name=remaining_quantity,json=remainingQuantity,proto3,customtype=cosmossdk.io/math.Int" json:"remaining_quantity"`
	// remaining_balance is remaining order balance.
	RemainingBalance cosmossdk_io_math.Int `protobuf:"bytes,7,opt,name=remaining_balance,json=remainingBalance,proto3,customtype=cosmossdk.io/math.Int" json:"remaining_balance"`
}

func (m *OrderBookStoreRecord) Reset()         { *m = OrderBookStoreRecord{} }
func (m *OrderBookStoreRecord) String() string { return proto.CompactTextString(m) }
func (*OrderBookStoreRecord) ProtoMessage()    {}
func (*OrderBookStoreRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_302bb6c9a553771c, []int{2}
}
func (m *OrderBookStoreRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OrderBookStoreRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OrderBookStoreRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OrderBookStoreRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OrderBookStoreRecord.Merge(m, src)
}
func (m *OrderBookStoreRecord) XXX_Size() int {
	return m.Size()
}
func (m *OrderBookStoreRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_OrderBookStoreRecord.DiscardUnknown(m)
}

var xxx_messageInfo_OrderBookStoreRecord proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("coreum.dex.v1.Side", Side_name, Side_value)
	proto.RegisterType((*Order)(nil), "coreum.dex.v1.Order")
	proto.RegisterType((*OrderBookRecord)(nil), "coreum.dex.v1.OrderBookRecord")
	proto.RegisterType((*OrderBookStoreRecord)(nil), "coreum.dex.v1.OrderBookStoreRecord")
}

func init() { proto.RegisterFile("coreum/dex/v1/order.proto", fileDescriptor_302bb6c9a553771c) }

var fileDescriptor_302bb6c9a553771c = []byte{
	// 453 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xb6, 0x5d, 0x3b, 0x3f, 0x8b, 0xda, 0x86, 0xa5, 0x80, 0xf9, 0x73, 0xab, 0x70, 0xa0, 0xea,
	0xc1, 0x56, 0x80, 0x27, 0x08, 0x11, 0x52, 0x10, 0x52, 0xc1, 0xb9, 0x71, 0x89, 0x9c, 0xdd, 0x21,
	0x5d, 0x35, 0xde, 0x71, 0xd6, 0xeb, 0x28, 0x79, 0x0b, 0x1e, 0x2b, 0xc7, 0x1e, 0x11, 0x87, 0x0a,
	0x12, 0x09, 0xf1, 0x18, 0xc8, 0x1b, 0x93, 0xc2, 0x05, 0x45, 0x88, 0xdb, 0xce, 0xf7, 0xcd, 0xb7,
	0x33, 0xdf, 0x68, 0x86, 0x3c, 0x60, 0xa8, 0xa0, 0x48, 0x23, 0x0e, 0xf3, 0x68, 0xd6, 0x89, 0x50,
	0x71, 0x50, 0x61, 0xa6, 0x50, 0x23, 0xdd, 0xdf, 0x50, 0x21, 0x87, 0x79, 0x38, 0xeb, 0x3c, 0x3c,
	0x1a, 0xe3, 0x18, 0x0d, 0x13, 0x95, 0xaf, 0x4d, 0x52, 0xfb, 0x3e, 0xf1, 0xce, 0x4b, 0x0d, 0x3d,
	0x20, 0x8e, 0xe0, 0xbe, 0x7d, 0x62, 0x9f, 0x36, 0x63, 0x47, 0xf0, 0xf6, 0x77, 0x87, 0x1c, 0x1a,
	0xa6, 0x8b, 0x78, 0x19, 0x03, 0x43, 0xc5, 0xe9, 0x3d, 0x52, 0xcb, 0x12, 0xa1, 0xfa, 0x3d, 0x93,
	0xe7, 0xc6, 0x55, 0x44, 0x9f, 0x11, 0x37, 0x17, 0x1c, 0x7c, 0xe7, 0xc4, 0x3e, 0x3d, 0x78, 0x7e,
	0x27, 0xfc, 0xa3, 0x70, 0x38, 0x10, 0x1c, 0x62, 0x93, 0x40, 0x9f, 0x12, 0x2f, 0x53, 0x82, 0x81,
	0xbf, 0x57, 0xd6, 0xe9, 0xee, 0x2f, 0xaf, 0x8f, 0xad, 0x2f, 0xd7, 0xc7, 0xde, 0xbb, 0x12, 0x8c,
	0x37, 0x1c, 0x7d, 0x44, 0x9a, 0xc6, 0xc6, 0x30, 0x87, 0xa9, 0xef, 0x9a, 0x42, 0x0d, 0x03, 0x0c,
	0x60, 0x4a, 0x7d, 0x52, 0x37, 0xef, 0x7e, 0xcf, 0xf7, 0x4c, 0xaf, 0xbf, 0x42, 0xfa, 0x98, 0x34,
	0x13, 0xc6, 0xb0, 0x90, 0xba, 0xdf, 0xf3, 0x6b, 0x86, 0xbb, 0x01, 0xe8, 0x5b, 0x42, 0x15, 0xa4,
	0x89, 0x90, 0x42, 0x8e, 0x87, 0xd3, 0x22, 0x91, 0x5a, 0xe8, 0x85, 0x5f, 0x37, 0x6d, 0x3c, 0xa9,
	0xda, 0xb8, 0xcb, 0x30, 0x4f, 0x31, 0xcf, 0xf9, 0x65, 0x28, 0x30, 0x4a, 0x13, 0x7d, 0x11, 0xf6,
	0xa5, 0x8e, 0x6f, 0x6f, 0x85, 0xef, 0x2b, 0x1d, 0x7d, 0x43, 0x6e, 0xc0, 0xe1, 0x28, 0x99, 0x24,
	0x92, 0x81, 0xdf, 0xd8, 0xe5, 0xb3, 0xd6, 0x56, 0xd7, 0xdd, 0xc8, 0xda, 0x3f, 0x6c, 0x72, 0xb4,
	0x1d, 0xf4, 0x40, 0xa3, 0x82, 0x6a, 0xda, 0xbf, 0x59, 0x75, 0xff, 0x62, 0xd5, 0xdb, 0xcd, 0x6a,
	0xed, 0x7f, 0x5a, 0xad, 0xff, 0x93, 0xd5, 0xb3, 0x33, 0xe2, 0x96, 0xcb, 0x40, 0x0f, 0xc9, 0xad,
	0x42, 0xe6, 0x19, 0x30, 0xf1, 0x51, 0x00, 0x6f, 0x59, 0xb4, 0x41, 0xdc, 0x1c, 0x26, 0x93, 0x96,
	0x4d, 0xeb, 0x64, 0x6f, 0x54, 0x2c, 0x5a, 0x4e, 0xf7, 0x7c, 0xf9, 0x2d, 0xb0, 0x96, 0xab, 0xc0,
	0xbe, 0x5a, 0x05, 0xf6, 0xd7, 0x55, 0x60, 0x7f, 0x5a, 0x07, 0xd6, 0xd5, 0x3a, 0xb0, 0x3e, 0xaf,
	0x03, 0xeb, 0x43, 0x67, 0x2c, 0xf4, 0x45, 0x31, 0x0a, 0x19, 0xa6, 0xd1, 0x2b, 0xb3, 0x6d, 0xaf,
	0xb1, 0x90, 0x3c, 0xd1, 0x02, 0x65, 0x54, 0x9d, 0xc4, 0xec, 0x65, 0x34, 0x37, 0x77, 0xa1, 0x17,
	0x19, 0xe4, 0xa3, 0x9a, 0x59, 0xf8, 0x17, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0xe7, 0x97, 0x63,
	0x6c, 0x32, 0x03, 0x00, 0x00,
}

func (m *Order) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Order) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Order) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintOrder(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *OrderBookRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OrderBookRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OrderBookRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.RemainingBalance.Size()
		i -= size
		if _, err := m.RemainingBalance.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOrder(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x42
	{
		size := m.RemainingQuantity.Size()
		i -= size
		if _, err := m.RemainingQuantity.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOrder(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	if len(m.AccountID) > 0 {
		i -= len(m.AccountID)
		copy(dAtA[i:], m.AccountID)
		i = encodeVarintOrder(dAtA, i, uint64(len(m.AccountID)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.OrderID) > 0 {
		i -= len(m.OrderID)
		copy(dAtA[i:], m.OrderID)
		i = encodeVarintOrder(dAtA, i, uint64(len(m.OrderID)))
		i--
		dAtA[i] = 0x2a
	}
	if m.OrderSeq != 0 {
		i = encodeVarintOrder(dAtA, i, uint64(m.OrderSeq))
		i--
		dAtA[i] = 0x20
	}
	{
		size := m.Price.Size()
		i -= size
		if _, err := m.Price.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOrder(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if m.Side != 0 {
		i = encodeVarintOrder(dAtA, i, uint64(m.Side))
		i--
		dAtA[i] = 0x10
	}
	if m.PairID != 0 {
		i = encodeVarintOrder(dAtA, i, uint64(m.PairID))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *OrderBookStoreRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OrderBookStoreRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OrderBookStoreRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.RemainingBalance.Size()
		i -= size
		if _, err := m.RemainingBalance.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOrder(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	{
		size := m.RemainingQuantity.Size()
		i -= size
		if _, err := m.RemainingQuantity.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOrder(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	if len(m.AccountID) > 0 {
		i -= len(m.AccountID)
		copy(dAtA[i:], m.AccountID)
		i = encodeVarintOrder(dAtA, i, uint64(len(m.AccountID)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.OrderID) > 0 {
		i -= len(m.OrderID)
		copy(dAtA[i:], m.OrderID)
		i = encodeVarintOrder(dAtA, i, uint64(len(m.OrderID)))
		i--
		dAtA[i] = 0x22
	}
	return len(dAtA) - i, nil
}

func encodeVarintOrder(dAtA []byte, offset int, v uint64) int {
	offset -= sovOrder(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Order) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovOrder(uint64(l))
	}
	return n
}

func (m *OrderBookRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PairID != 0 {
		n += 1 + sovOrder(uint64(m.PairID))
	}
	if m.Side != 0 {
		n += 1 + sovOrder(uint64(m.Side))
	}
	l = m.Price.Size()
	n += 1 + l + sovOrder(uint64(l))
	if m.OrderSeq != 0 {
		n += 1 + sovOrder(uint64(m.OrderSeq))
	}
	l = len(m.OrderID)
	if l > 0 {
		n += 1 + l + sovOrder(uint64(l))
	}
	l = len(m.AccountID)
	if l > 0 {
		n += 1 + l + sovOrder(uint64(l))
	}
	l = m.RemainingQuantity.Size()
	n += 1 + l + sovOrder(uint64(l))
	l = m.RemainingBalance.Size()
	n += 1 + l + sovOrder(uint64(l))
	return n
}

func (m *OrderBookStoreRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.OrderID)
	if l > 0 {
		n += 1 + l + sovOrder(uint64(l))
	}
	l = len(m.AccountID)
	if l > 0 {
		n += 1 + l + sovOrder(uint64(l))
	}
	l = m.RemainingQuantity.Size()
	n += 1 + l + sovOrder(uint64(l))
	l = m.RemainingBalance.Size()
	n += 1 + l + sovOrder(uint64(l))
	return n
}

func sovOrder(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozOrder(x uint64) (n int) {
	return sovOrder(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Order) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOrder
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
			return fmt.Errorf("proto: Order: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Order: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOrder(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOrder
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
func (m *OrderBookRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOrder
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
			return fmt.Errorf("proto: OrderBookRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OrderBookRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PairID", wireType)
			}
			m.PairID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PairID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Side", wireType)
			}
			m.Side = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Side |= Side(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Price", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Price.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrderSeq", wireType)
			}
			m.OrderSeq = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.OrderSeq |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrderID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrderID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainingQuantity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RemainingQuantity.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainingBalance", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RemainingBalance.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOrder(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOrder
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
func (m *OrderBookStoreRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOrder
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
			return fmt.Errorf("proto: OrderBookStoreRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OrderBookStoreRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrderID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrderID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainingQuantity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RemainingQuantity.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainingBalance", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOrder
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
				return ErrInvalidLengthOrder
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOrder
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RemainingBalance.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOrder(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOrder
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
func skipOrder(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowOrder
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
					return 0, ErrIntOverflowOrder
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
					return 0, ErrIntOverflowOrder
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
				return 0, ErrInvalidLengthOrder
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupOrder
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthOrder
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthOrder        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowOrder          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupOrder = fmt.Errorf("proto: unexpected end of group")
)
