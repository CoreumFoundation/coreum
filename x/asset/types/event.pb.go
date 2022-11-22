// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/v1/event.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
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

// EventFungibleTokenIssued is emitted on MsgIssueFungibleToken.
type EventFungibleTokenIssued struct {
	Denom         string                                 `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Issuer        string                                 `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Symbol        string                                 `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	SubUnit       string                                 `protobuf:"bytes,4,opt,name=sub_unit,json=subUnit,proto3" json:"sub_unit,omitempty"`
	Precision     uint32                                 `protobuf:"varint,5,opt,name=precision,proto3" json:"precision,omitempty"`
	Recipient     string                                 `protobuf:"bytes,6,opt,name=recipient,proto3" json:"recipient,omitempty"`
	InitialAmount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,7,opt,name=initial_amount,json=initialAmount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"initial_amount"`
	Description   string                                 `protobuf:"bytes,8,opt,name=description,proto3" json:"description,omitempty"`
	Features      []FungibleTokenFeature                 `protobuf:"varint,9,rep,packed,name=features,proto3,enum=coreum.asset.v1.FungibleTokenFeature" json:"features,omitempty"`
}

func (m *EventFungibleTokenIssued) Reset()         { *m = EventFungibleTokenIssued{} }
func (m *EventFungibleTokenIssued) String() string { return proto.CompactTextString(m) }
func (*EventFungibleTokenIssued) ProtoMessage()    {}
func (*EventFungibleTokenIssued) Descriptor() ([]byte, []int) {
	return fileDescriptor_aede4b64fdc52aa3, []int{0}
}
func (m *EventFungibleTokenIssued) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventFungibleTokenIssued) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventFungibleTokenIssued.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventFungibleTokenIssued) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventFungibleTokenIssued.Merge(m, src)
}
func (m *EventFungibleTokenIssued) XXX_Size() int {
	return m.Size()
}
func (m *EventFungibleTokenIssued) XXX_DiscardUnknown() {
	xxx_messageInfo_EventFungibleTokenIssued.DiscardUnknown(m)
}

var xxx_messageInfo_EventFungibleTokenIssued proto.InternalMessageInfo

func (m *EventFungibleTokenIssued) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetIssuer() string {
	if m != nil {
		return m.Issuer
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetSubUnit() string {
	if m != nil {
		return m.SubUnit
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetPrecision() uint32 {
	if m != nil {
		return m.Precision
	}
	return 0
}

func (m *EventFungibleTokenIssued) GetRecipient() string {
	if m != nil {
		return m.Recipient
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetFeatures() []FungibleTokenFeature {
	if m != nil {
		return m.Features
	}
	return nil
}

type EventFungibleTokenFrozen struct {
	Account string     `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Coin    types.Coin `protobuf:"bytes,2,opt,name=coin,proto3" json:"coin"`
}

func (m *EventFungibleTokenFrozen) Reset()         { *m = EventFungibleTokenFrozen{} }
func (m *EventFungibleTokenFrozen) String() string { return proto.CompactTextString(m) }
func (*EventFungibleTokenFrozen) ProtoMessage()    {}
func (*EventFungibleTokenFrozen) Descriptor() ([]byte, []int) {
	return fileDescriptor_aede4b64fdc52aa3, []int{1}
}
func (m *EventFungibleTokenFrozen) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventFungibleTokenFrozen) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventFungibleTokenFrozen.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventFungibleTokenFrozen) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventFungibleTokenFrozen.Merge(m, src)
}
func (m *EventFungibleTokenFrozen) XXX_Size() int {
	return m.Size()
}
func (m *EventFungibleTokenFrozen) XXX_DiscardUnknown() {
	xxx_messageInfo_EventFungibleTokenFrozen.DiscardUnknown(m)
}

var xxx_messageInfo_EventFungibleTokenFrozen proto.InternalMessageInfo

func (m *EventFungibleTokenFrozen) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *EventFungibleTokenFrozen) GetCoin() types.Coin {
	if m != nil {
		return m.Coin
	}
	return types.Coin{}
}

type EventFungibleTokenUnfrozen struct {
	Account string     `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Coin    types.Coin `protobuf:"bytes,2,opt,name=coin,proto3" json:"coin"`
}

func (m *EventFungibleTokenUnfrozen) Reset()         { *m = EventFungibleTokenUnfrozen{} }
func (m *EventFungibleTokenUnfrozen) String() string { return proto.CompactTextString(m) }
func (*EventFungibleTokenUnfrozen) ProtoMessage()    {}
func (*EventFungibleTokenUnfrozen) Descriptor() ([]byte, []int) {
	return fileDescriptor_aede4b64fdc52aa3, []int{2}
}
func (m *EventFungibleTokenUnfrozen) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventFungibleTokenUnfrozen) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventFungibleTokenUnfrozen.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventFungibleTokenUnfrozen) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventFungibleTokenUnfrozen.Merge(m, src)
}
func (m *EventFungibleTokenUnfrozen) XXX_Size() int {
	return m.Size()
}
func (m *EventFungibleTokenUnfrozen) XXX_DiscardUnknown() {
	xxx_messageInfo_EventFungibleTokenUnfrozen.DiscardUnknown(m)
}

var xxx_messageInfo_EventFungibleTokenUnfrozen proto.InternalMessageInfo

func (m *EventFungibleTokenUnfrozen) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *EventFungibleTokenUnfrozen) GetCoin() types.Coin {
	if m != nil {
		return m.Coin
	}
	return types.Coin{}
}

func init() {
	proto.RegisterType((*EventFungibleTokenIssued)(nil), "coreum.asset.v1.EventFungibleTokenIssued")
	proto.RegisterType((*EventFungibleTokenFrozen)(nil), "coreum.asset.v1.EventFungibleTokenFrozen")
	proto.RegisterType((*EventFungibleTokenUnfrozen)(nil), "coreum.asset.v1.EventFungibleTokenUnfrozen")
}

func init() { proto.RegisterFile("coreum/asset/v1/event.proto", fileDescriptor_aede4b64fdc52aa3) }

var fileDescriptor_aede4b64fdc52aa3 = []byte{
	// 461 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x52, 0xcd, 0x6e, 0x13, 0x3d,
	0x14, 0xcd, 0xb4, 0x69, 0x7e, 0x5c, 0xb5, 0x9f, 0x64, 0x55, 0x9f, 0xdc, 0x80, 0xa6, 0x51, 0x24,
	0x50, 0x36, 0xd8, 0xa4, 0x7d, 0x82, 0xa6, 0x22, 0x52, 0x59, 0x8e, 0xc8, 0x86, 0x4d, 0x35, 0x3f,
	0xb7, 0xc1, 0x4a, 0xc6, 0x77, 0x34, 0xb6, 0x23, 0xca, 0x53, 0xf0, 0x58, 0x5d, 0x76, 0x89, 0x58,
	0x54, 0x28, 0x79, 0x04, 0x5e, 0x00, 0xd9, 0x1e, 0x68, 0xa1, 0x6c, 0x59, 0x8d, 0xef, 0x39, 0xd7,
	0xf7, 0x8c, 0xcf, 0xb9, 0xe4, 0x59, 0x8e, 0x35, 0xd8, 0x52, 0xa4, 0x5a, 0x83, 0x11, 0xeb, 0x89,
	0x80, 0x35, 0x28, 0xc3, 0xab, 0x1a, 0x0d, 0xd2, 0xff, 0x02, 0xc9, 0x3d, 0xc9, 0xd7, 0x93, 0xc1,
	0xd1, 0x02, 0x17, 0xe8, 0x39, 0xe1, 0x4e, 0xa1, 0x6d, 0x10, 0xe7, 0xa8, 0x4b, 0xd4, 0x22, 0x4b,
	0x35, 0x88, 0xf5, 0x24, 0x03, 0x93, 0x4e, 0x44, 0x8e, 0x52, 0x35, 0xfc, 0x13, 0x8d, 0x30, 0xcf,
	0x93, 0xa3, 0xef, 0x3b, 0x84, 0xbd, 0x71, 0x9a, 0x33, 0xab, 0x16, 0x32, 0x5b, 0xc1, 0x3b, 0x5c,
	0x82, 0xba, 0xd4, 0xda, 0x42, 0x41, 0x8f, 0xc8, 0x5e, 0x01, 0x0a, 0x4b, 0x16, 0x0d, 0xa3, 0x71,
	0x3f, 0x09, 0x05, 0xfd, 0x9f, 0x74, 0xa4, 0xe3, 0x6b, 0xb6, 0xe3, 0xe1, 0xa6, 0x72, 0xb8, 0xbe,
	0x29, 0x33, 0x5c, 0xb1, 0xdd, 0x80, 0x87, 0x8a, 0x1e, 0x93, 0x9e, 0xb6, 0xd9, 0x95, 0x55, 0xd2,
	0xb0, 0xb6, 0x67, 0xba, 0xda, 0x66, 0x73, 0x25, 0x0d, 0x7d, 0x4e, 0xfa, 0x55, 0x0d, 0xb9, 0xd4,
	0x12, 0x15, 0xdb, 0x1b, 0x46, 0xe3, 0x83, 0xe4, 0x01, 0x70, 0xac, 0x3b, 0x57, 0x12, 0x94, 0x61,
	0x1d, 0x7f, 0xf3, 0x01, 0xa0, 0x73, 0x72, 0x28, 0x95, 0x34, 0x32, 0x5d, 0x5d, 0xa5, 0x25, 0x5a,
	0x65, 0x58, 0xd7, 0xb5, 0x4c, 0xf9, 0xed, 0xfd, 0x49, 0xeb, 0xeb, 0xfd, 0xc9, 0xcb, 0x85, 0x34,
	0x1f, 0x6c, 0xc6, 0x73, 0x2c, 0x45, 0xe3, 0x50, 0xf8, 0xbc, 0xd2, 0xc5, 0x52, 0x98, 0x9b, 0x0a,
	0x34, 0xbf, 0x54, 0x26, 0x39, 0x68, 0xa6, 0x9c, 0xfb, 0x21, 0x74, 0x48, 0xf6, 0x0b, 0xd0, 0x79,
	0x2d, 0x2b, 0xe3, 0x7e, 0xaa, 0xe7, 0x65, 0x1f, 0x43, 0xf4, 0x9c, 0xf4, 0xae, 0x21, 0x35, 0xb6,
	0x06, 0xcd, 0xfa, 0xc3, 0xdd, 0xf1, 0xe1, 0xe9, 0x0b, 0xfe, 0x47, 0x52, 0xfc, 0x37, 0x37, 0x67,
	0xa1, 0x3b, 0xf9, 0x75, 0x6d, 0x24, 0xff, 0x66, 0xfa, 0xac, 0xc6, 0x4f, 0xa0, 0x28, 0x23, 0xdd,
	0x34, 0xcf, 0xfd, 0x83, 0x82, 0xed, 0x3f, 0x4b, 0x7a, 0x46, 0xda, 0x2e, 0x56, 0x6f, 0xfb, 0xfe,
	0xe9, 0x31, 0x0f, 0xcf, 0xe1, 0x2e, 0x77, 0xde, 0xe4, 0xce, 0x2f, 0x50, 0xaa, 0x69, 0xdb, 0x59,
	0x90, 0xf8, 0xe6, 0xd1, 0x92, 0x0c, 0x9e, 0x4a, 0xcd, 0xd5, 0xf5, 0xbf, 0x10, 0x9b, 0xbe, 0xbd,
	0xdd, 0xc4, 0xd1, 0xdd, 0x26, 0x8e, 0xbe, 0x6d, 0xe2, 0xe8, 0xf3, 0x36, 0x6e, 0xdd, 0x6d, 0xe3,
	0xd6, 0x97, 0x6d, 0xdc, 0x7a, 0xff, 0xfa, 0x51, 0x1a, 0x17, 0xde, 0xac, 0x19, 0x5a, 0x55, 0xa4,
	0xce, 0x51, 0xd1, 0x2c, 0xe8, 0xc7, 0x66, 0x45, 0x7d, 0x36, 0x59, 0xc7, 0x2f, 0xe8, 0xd9, 0x8f,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x9c, 0x4b, 0x57, 0x43, 0x23, 0x03, 0x00, 0x00,
}

func (m *EventFungibleTokenIssued) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventFungibleTokenIssued) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventFungibleTokenIssued) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Features) > 0 {
		dAtA2 := make([]byte, len(m.Features)*10)
		var j1 int
		for _, num := range m.Features {
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintEvent(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x42
	}
	{
		size := m.InitialAmount.Size()
		i -= size
		if _, err := m.InitialAmount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	if len(m.Recipient) > 0 {
		i -= len(m.Recipient)
		copy(dAtA[i:], m.Recipient)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Recipient)))
		i--
		dAtA[i] = 0x32
	}
	if m.Precision != 0 {
		i = encodeVarintEvent(dAtA, i, uint64(m.Precision))
		i--
		dAtA[i] = 0x28
	}
	if len(m.SubUnit) > 0 {
		i -= len(m.SubUnit)
		copy(dAtA[i:], m.SubUnit)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.SubUnit)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Symbol) > 0 {
		i -= len(m.Symbol)
		copy(dAtA[i:], m.Symbol)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Symbol)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Issuer) > 0 {
		i -= len(m.Issuer)
		copy(dAtA[i:], m.Issuer)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Issuer)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventFungibleTokenFrozen) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventFungibleTokenFrozen) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventFungibleTokenFrozen) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Coin.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Account) > 0 {
		i -= len(m.Account)
		copy(dAtA[i:], m.Account)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Account)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventFungibleTokenUnfrozen) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventFungibleTokenUnfrozen) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventFungibleTokenUnfrozen) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Coin.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Account) > 0 {
		i -= len(m.Account)
		copy(dAtA[i:], m.Account)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Account)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintEvent(dAtA []byte, offset int, v uint64) int {
	offset -= sovEvent(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EventFungibleTokenIssued) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.Issuer)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.Symbol)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.SubUnit)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	if m.Precision != 0 {
		n += 1 + sovEvent(uint64(m.Precision))
	}
	l = len(m.Recipient)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = m.InitialAmount.Size()
	n += 1 + l + sovEvent(uint64(l))
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	if len(m.Features) > 0 {
		l = 0
		for _, e := range m.Features {
			l += sovEvent(uint64(e))
		}
		n += 1 + sovEvent(uint64(l)) + l
	}
	return n
}

func (m *EventFungibleTokenFrozen) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Account)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = m.Coin.Size()
	n += 1 + l + sovEvent(uint64(l))
	return n
}

func (m *EventFungibleTokenUnfrozen) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Account)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = m.Coin.Size()
	n += 1 + l + sovEvent(uint64(l))
	return n
}

func sovEvent(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEvent(x uint64) (n int) {
	return sovEvent(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventFungibleTokenIssued) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvent
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
			return fmt.Errorf("proto: EventFungibleTokenIssued: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventFungibleTokenIssued: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Issuer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Issuer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Symbol", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Symbol = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SubUnit", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SubUnit = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Precision", wireType)
			}
			m.Precision = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Precision |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Recipient", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Recipient = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InitialAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.InitialAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType == 0 {
				var v FungibleTokenFeature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowEvent
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= FungibleTokenFeature(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Features = append(m.Features, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowEvent
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthEvent
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthEvent
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Features) == 0 {
					m.Features = make([]FungibleTokenFeature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v FungibleTokenFeature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowEvent
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= FungibleTokenFeature(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Features = append(m.Features, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Features", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEvent(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvent
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
func (m *EventFungibleTokenFrozen) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvent
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
			return fmt.Errorf("proto: EventFungibleTokenFrozen: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventFungibleTokenFrozen: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Account", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Account = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Coin", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Coin.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvent(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvent
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
func (m *EventFungibleTokenUnfrozen) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvent
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
			return fmt.Errorf("proto: EventFungibleTokenUnfrozen: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventFungibleTokenUnfrozen: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Account", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Account = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Coin", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Coin.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvent(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvent
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
func skipEvent(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvent
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
					return 0, ErrIntOverflowEvent
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
					return 0, ErrIntOverflowEvent
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
				return 0, ErrInvalidLengthEvent
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupEvent
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthEvent
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthEvent        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvent          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupEvent = fmt.Errorf("proto: unexpected end of group")
)
