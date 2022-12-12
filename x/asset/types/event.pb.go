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
	Subunit       string                                 `protobuf:"bytes,4,opt,name=subunit,proto3" json:"subunit,omitempty"`
	Precision     uint32                                 `protobuf:"varint,5,opt,name=precision,proto3" json:"precision,omitempty"`
	InitialAmount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,6,opt,name=initial_amount,json=initialAmount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"initial_amount"`
	Description   string                                 `protobuf:"bytes,7,opt,name=description,proto3" json:"description,omitempty"`
	Features      []FungibleTokenFeature                 `protobuf:"varint,8,rep,packed,name=features,proto3,enum=coreum.asset.v1.FungibleTokenFeature" json:"features,omitempty"`
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

func (m *EventFungibleTokenIssued) GetSubunit() string {
	if m != nil {
		return m.Subunit
	}
	return ""
}

func (m *EventFungibleTokenIssued) GetPrecision() uint32 {
	if m != nil {
		return m.Precision
	}
	return 0
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

type EventFungibleTokenFrozenAmountChanged struct {
	Account        string     `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	PreviousAmount types.Coin `protobuf:"bytes,2,opt,name=previous_amount,json=previousAmount,proto3" json:"previous_amount"`
	CurrentAmount  types.Coin `protobuf:"bytes,3,opt,name=current_amount,json=currentAmount,proto3" json:"current_amount"`
}

func (m *EventFungibleTokenFrozenAmountChanged) Reset()         { *m = EventFungibleTokenFrozenAmountChanged{} }
func (m *EventFungibleTokenFrozenAmountChanged) String() string { return proto.CompactTextString(m) }
func (*EventFungibleTokenFrozenAmountChanged) ProtoMessage()    {}
func (*EventFungibleTokenFrozenAmountChanged) Descriptor() ([]byte, []int) {
	return fileDescriptor_aede4b64fdc52aa3, []int{1}
}
func (m *EventFungibleTokenFrozenAmountChanged) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventFungibleTokenFrozenAmountChanged) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventFungibleTokenFrozenAmountChanged.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventFungibleTokenFrozenAmountChanged) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventFungibleTokenFrozenAmountChanged.Merge(m, src)
}
func (m *EventFungibleTokenFrozenAmountChanged) XXX_Size() int {
	return m.Size()
}
func (m *EventFungibleTokenFrozenAmountChanged) XXX_DiscardUnknown() {
	xxx_messageInfo_EventFungibleTokenFrozenAmountChanged.DiscardUnknown(m)
}

var xxx_messageInfo_EventFungibleTokenFrozenAmountChanged proto.InternalMessageInfo

func (m *EventFungibleTokenFrozenAmountChanged) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *EventFungibleTokenFrozenAmountChanged) GetPreviousAmount() types.Coin {
	if m != nil {
		return m.PreviousAmount
	}
	return types.Coin{}
}

func (m *EventFungibleTokenFrozenAmountChanged) GetCurrentAmount() types.Coin {
	if m != nil {
		return m.CurrentAmount
	}
	return types.Coin{}
}

func init() {
	proto.RegisterType((*EventFungibleTokenIssued)(nil), "coreum.asset.v1.EventFungibleTokenIssued")
	proto.RegisterType((*EventFungibleTokenFrozenAmountChanged)(nil), "coreum.asset.v1.EventFungibleTokenFrozenAmountChanged")
}

func init() { proto.RegisterFile("coreum/asset/v1/event.proto", fileDescriptor_aede4b64fdc52aa3) }

var fileDescriptor_aede4b64fdc52aa3 = []byte{
	// 467 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xc7, 0xe3, 0x7e, 0xa4, 0xed, 0x56, 0x49, 0x25, 0xab, 0x42, 0xa6, 0x20, 0x37, 0xaa, 0x54,
	0x94, 0x0b, 0xbb, 0xa4, 0x3c, 0x41, 0x13, 0x61, 0x51, 0x8e, 0x16, 0x5c, 0xb8, 0x20, 0x7f, 0x0c,
	0xee, 0xaa, 0xf1, 0x8e, 0xb5, 0x1f, 0x16, 0xe5, 0x29, 0x78, 0xac, 0x1e, 0x2b, 0xc1, 0x01, 0x71,
	0xa8, 0x50, 0xf2, 0x22, 0x68, 0x77, 0x6d, 0x28, 0xf4, 0xc2, 0xc9, 0x9e, 0xf9, 0xef, 0xfc, 0x77,
	0xe7, 0x37, 0x43, 0x9e, 0x14, 0x28, 0xc1, 0xd4, 0x2c, 0x53, 0x0a, 0x34, 0x6b, 0x67, 0x0c, 0x5a,
	0x10, 0x9a, 0x36, 0x12, 0x35, 0x86, 0x07, 0x5e, 0xa4, 0x4e, 0xa4, 0xed, 0xec, 0xe8, 0xb0, 0xc2,
	0x0a, 0x9d, 0xc6, 0xec, 0x9f, 0x3f, 0x76, 0x14, 0x17, 0xa8, 0x6a, 0x54, 0x2c, 0xcf, 0x14, 0xb0,
	0x76, 0x96, 0x83, 0xce, 0x66, 0xac, 0x40, 0x2e, 0x3a, 0xfd, 0xc1, 0x1d, 0xde, 0xcf, 0x89, 0x27,
	0xdf, 0x36, 0x48, 0xf4, 0xca, 0xde, 0x99, 0x18, 0x51, 0xf1, 0x7c, 0x09, 0x6f, 0xf1, 0x0a, 0xc4,
	0x85, 0x52, 0x06, 0xca, 0xf0, 0x90, 0x6c, 0x97, 0x20, 0xb0, 0x8e, 0x82, 0x49, 0x30, 0xdd, 0x4b,
	0x7d, 0x10, 0x3e, 0x22, 0x43, 0x6e, 0x75, 0x19, 0x6d, 0xb8, 0x74, 0x17, 0xd9, 0xbc, 0xba, 0xae,
	0x73, 0x5c, 0x46, 0x9b, 0x3e, 0xef, 0xa3, 0x30, 0x22, 0x3b, 0xca, 0xe4, 0x46, 0x70, 0x1d, 0x6d,
	0x39, 0xa1, 0x0f, 0xc3, 0xa7, 0x64, 0xaf, 0x91, 0x50, 0x70, 0xc5, 0x51, 0x44, 0xdb, 0x93, 0x60,
	0x3a, 0x4a, 0xff, 0x24, 0xc2, 0x77, 0x64, 0xcc, 0x05, 0xd7, 0x3c, 0x5b, 0x7e, 0xc8, 0x6a, 0x34,
	0x42, 0x47, 0x43, 0x5b, 0x3e, 0xa7, 0x37, 0x77, 0xc7, 0x83, 0x1f, 0x77, 0xc7, 0xcf, 0x2a, 0xae,
	0x2f, 0x4d, 0x4e, 0x0b, 0xac, 0x59, 0x87, 0xc0, 0x7f, 0x9e, 0xab, 0xf2, 0x8a, 0xe9, 0xeb, 0x06,
	0x14, 0xbd, 0x10, 0x3a, 0x1d, 0x75, 0x2e, 0xe7, 0xce, 0x24, 0x9c, 0x90, 0xfd, 0x12, 0x54, 0x21,
	0x79, 0xa3, 0xed, 0xb5, 0x3b, 0xee, 0x49, 0xf7, 0x53, 0xe1, 0x39, 0xd9, 0xfd, 0x08, 0x99, 0x36,
	0x12, 0x54, 0xb4, 0x3b, 0xd9, 0x9c, 0x8e, 0xcf, 0x4e, 0xe9, 0x3f, 0xa3, 0xa0, 0x7f, 0xe1, 0x4a,
	0xfc, 0xe9, 0xf4, 0x77, 0xd9, 0xc9, 0xd7, 0x80, 0x9c, 0x3e, 0xc4, 0x9a, 0x48, 0xfc, 0x0c, 0xc2,
	0xbf, 0x63, 0x71, 0x99, 0x89, 0x0a, 0x4a, 0x4b, 0x27, 0x2b, 0x0a, 0xd7, 0x9e, 0xa7, 0xdc, 0x87,
	0xe1, 0x6b, 0x72, 0xd0, 0x48, 0x68, 0x39, 0x1a, 0xd5, 0x03, 0xb0, 0xc0, 0xf7, 0xcf, 0x1e, 0x53,
	0xdf, 0x27, 0xb5, 0x13, 0xa7, 0xdd, 0xc4, 0xe9, 0x02, 0xb9, 0x98, 0x6f, 0x59, 0x36, 0xe9, 0xb8,
	0xaf, 0xeb, 0x5a, 0x4e, 0xc8, 0xb8, 0x30, 0x52, 0x82, 0xd0, 0xbd, 0xd1, 0xe6, 0xff, 0x19, 0x8d,
	0xba, 0x32, 0xef, 0x33, 0x7f, 0x73, 0xb3, 0x8a, 0x83, 0xdb, 0x55, 0x1c, 0xfc, 0x5c, 0xc5, 0xc1,
	0x97, 0x75, 0x3c, 0xb8, 0x5d, 0xc7, 0x83, 0xef, 0xeb, 0x78, 0xf0, 0xfe, 0xc5, 0xbd, 0x59, 0x2c,
	0x1c, 0xaa, 0x04, 0x8d, 0x28, 0x33, 0xcb, 0x93, 0x75, 0xfb, 0xf7, 0xa9, 0xdb, 0x40, 0x37, 0x99,
	0x7c, 0xe8, 0xf6, 0xef, 0xe5, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x68, 0xad, 0xf3, 0xd1, 0x02,
	0x03, 0x00, 0x00,
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
		dAtA[i] = 0x42
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x3a
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
	dAtA[i] = 0x32
	if m.Precision != 0 {
		i = encodeVarintEvent(dAtA, i, uint64(m.Precision))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Subunit) > 0 {
		i -= len(m.Subunit)
		copy(dAtA[i:], m.Subunit)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Subunit)))
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

func (m *EventFungibleTokenFrozenAmountChanged) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventFungibleTokenFrozenAmountChanged) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventFungibleTokenFrozenAmountChanged) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.CurrentAmount.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.PreviousAmount.MarshalToSizedBuffer(dAtA[:i])
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
	l = len(m.Subunit)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	if m.Precision != 0 {
		n += 1 + sovEvent(uint64(m.Precision))
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

func (m *EventFungibleTokenFrozenAmountChanged) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Account)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = m.PreviousAmount.Size()
	n += 1 + l + sovEvent(uint64(l))
	l = m.CurrentAmount.Size()
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
				return fmt.Errorf("proto: wrong wireType = %d for field Subunit", wireType)
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
			m.Subunit = string(dAtA[iNdEx:postIndex])
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
		case 7:
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
		case 8:
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
func (m *EventFungibleTokenFrozenAmountChanged) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: EventFungibleTokenFrozenAmountChanged: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventFungibleTokenFrozenAmountChanged: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field PreviousAmount", wireType)
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
			if err := m.PreviousAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentAmount", wireType)
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
			if err := m.CurrentAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
