// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/ft/v1/token.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
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

// Feature defines possible features of fungible token.
type Feature int32

const (
	Feature_minting      Feature = 0
	Feature_burning      Feature = 1
	Feature_freezing     Feature = 2
	Feature_whitelisting Feature = 3
)

var Feature_name = map[int32]string{
	0: "minting",
	1: "burning",
	2: "freezing",
	3: "whitelisting",
}

var Feature_value = map[string]int32{
	"minting":      0,
	"burning":      1,
	"freezing":     2,
	"whitelisting": 3,
}

func (x Feature) String() string {
	return proto.EnumName(Feature_name, int32(x))
}

func (Feature) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fe80c7a2c55589e7, []int{0}
}

// Definition defines the fungible token settings to store.
type Definition struct {
	Denom    string    `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Issuer   string    `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Features []Feature `protobuf:"varint,3,rep,packed,name=features,proto3,enum=coreum.asset.ft.v1.Feature" json:"features,omitempty"`
	// burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine
	// burn_amount. This value will be burnt on top of the send amount.
	BurnRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,4,opt,name=burn_rate,json=burnRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"burn_rate"`
	// send_commission_rate is a number between 0 and 1 which will be multiplied by send amount to determine
	// amount sent to the token issuer account.
	SendCommissionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,5,opt,name=send_commission_rate,json=sendCommissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"send_commission_rate"`
}

func (m *Definition) Reset()         { *m = Definition{} }
func (m *Definition) String() string { return proto.CompactTextString(m) }
func (*Definition) ProtoMessage()    {}
func (*Definition) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe80c7a2c55589e7, []int{0}
}
func (m *Definition) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Definition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Definition.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Definition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Definition.Merge(m, src)
}
func (m *Definition) XXX_Size() int {
	return m.Size()
}
func (m *Definition) XXX_DiscardUnknown() {
	xxx_messageInfo_Definition.DiscardUnknown(m)
}

var xxx_messageInfo_Definition proto.InternalMessageInfo

// Token is a full representation of the fungible token.
type Token struct {
	Denom          string    `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Issuer         string    `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Symbol         string    `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Subunit        string    `protobuf:"bytes,4,opt,name=subunit,proto3" json:"subunit,omitempty"`
	Precision      uint32    `protobuf:"varint,5,opt,name=precision,proto3" json:"precision,omitempty"`
	Description    string    `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	GloballyFrozen bool      `protobuf:"varint,7,opt,name=globally_frozen,json=globallyFrozen,proto3" json:"globally_frozen,omitempty"`
	Features       []Feature `protobuf:"varint,8,rep,packed,name=features,proto3,enum=coreum.asset.ft.v1.Feature" json:"features,omitempty"`
	// burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine
	// burn_amount. This value will be burnt on top of the send amount.
	BurnRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,9,opt,name=burn_rate,json=burnRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"burn_rate"`
	// send_commission_rate is a number between 0 and 1 which will be multiplied by send amount to determine
	// amount sent to the token issuer account.
	SendCommissionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,10,opt,name=send_commission_rate,json=sendCommissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"send_commission_rate"`
}

func (m *Token) Reset()         { *m = Token{} }
func (m *Token) String() string { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()    {}
func (*Token) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe80c7a2c55589e7, []int{1}
}
func (m *Token) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Token) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Token.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Token) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Token.Merge(m, src)
}
func (m *Token) XXX_Size() int {
	return m.Size()
}
func (m *Token) XXX_DiscardUnknown() {
	xxx_messageInfo_Token.DiscardUnknown(m)
}

var xxx_messageInfo_Token proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("coreum.asset.ft.v1.Feature", Feature_name, Feature_value)
	proto.RegisterType((*Definition)(nil), "coreum.asset.ft.v1.Definition")
	proto.RegisterType((*Token)(nil), "coreum.asset.ft.v1.Token")
}

func init() { proto.RegisterFile("coreum/asset/ft/v1/token.proto", fileDescriptor_fe80c7a2c55589e7) }

var fileDescriptor_fe80c7a2c55589e7 = []byte{
	// 500 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x53, 0xcb, 0x6e, 0xd3, 0x40,
	0x14, 0xb5, 0x63, 0x92, 0x38, 0xd3, 0x52, 0xaa, 0x51, 0x54, 0x59, 0x05, 0x39, 0x51, 0x17, 0x10,
	0x21, 0xe1, 0x51, 0xe8, 0x02, 0x89, 0x65, 0x53, 0x65, 0x83, 0xd8, 0x58, 0xac, 0xd8, 0x04, 0x3f,
	0xae, 0xdd, 0x51, 0xed, 0x99, 0x68, 0x66, 0x1c, 0x48, 0xbf, 0x80, 0x25, 0x9f, 0xd0, 0x0f, 0xe0,
	0x2b, 0x58, 0x75, 0xd9, 0x25, 0x62, 0x51, 0xa1, 0x64, 0xc3, 0x67, 0xa0, 0x19, 0xbb, 0x0f, 0xc4,
	0x8a, 0x88, 0xae, 0xec, 0x73, 0xee, 0xf5, 0xf1, 0xb9, 0xf7, 0xe8, 0x22, 0x3f, 0xe1, 0x02, 0xaa,
	0x92, 0x44, 0x52, 0x82, 0x22, 0x99, 0x22, 0x8b, 0x31, 0x51, 0xfc, 0x14, 0x58, 0x30, 0x17, 0x5c,
	0x71, 0x8c, 0xeb, 0x7a, 0x60, 0xea, 0x41, 0xa6, 0x82, 0xc5, 0x78, 0xbf, 0x9f, 0xf3, 0x9c, 0x9b,
	0x32, 0xd1, 0x6f, 0x75, 0xe7, 0xbe, 0x9f, 0x70, 0x59, 0x72, 0x49, 0xe2, 0x48, 0x02, 0x59, 0x8c,
	0x63, 0x50, 0xd1, 0x98, 0x24, 0x9c, 0x36, 0x4a, 0x07, 0x5f, 0x5b, 0x08, 0x1d, 0x43, 0x46, 0x19,
	0x55, 0x94, 0x33, 0xdc, 0x47, 0xed, 0x14, 0x18, 0x2f, 0x3d, 0x7b, 0x68, 0x8f, 0x7a, 0x61, 0x0d,
	0xf0, 0x1e, 0xea, 0x50, 0x29, 0x2b, 0x10, 0x5e, 0xcb, 0xd0, 0x0d, 0xc2, 0xaf, 0x90, 0x9b, 0x41,
	0xa4, 0x2a, 0x01, 0xd2, 0x73, 0x86, 0xce, 0x68, 0xe7, 0xe5, 0xe3, 0xe0, 0x6f, 0x67, 0xc1, 0xb4,
	0xee, 0x09, 0x6f, 0x9a, 0xf1, 0x1b, 0xd4, 0x8b, 0x2b, 0xc1, 0x66, 0x22, 0x52, 0xe0, 0x3d, 0xd0,
	0x9a, 0x47, 0xc1, 0xc5, 0xd5, 0xc0, 0xfa, 0x71, 0x35, 0x78, 0x9a, 0x53, 0x75, 0x52, 0xc5, 0x41,
	0xc2, 0x4b, 0xd2, 0x78, 0xaf, 0x1f, 0x2f, 0x64, 0x7a, 0x4a, 0xd4, 0x72, 0x0e, 0x32, 0x38, 0x86,
	0x24, 0x74, 0xb5, 0x40, 0x18, 0x29, 0xc0, 0x1f, 0x50, 0x5f, 0x02, 0x4b, 0x67, 0x09, 0x2f, 0x4b,
	0x2a, 0x25, 0xe5, 0x8d, 0x6e, 0x7b, 0x23, 0x5d, 0xac, 0xb5, 0x26, 0x37, 0x52, 0xfa, 0x0f, 0xaf,
	0xdd, 0xcf, 0xe7, 0x03, 0xeb, 0xd7, 0xf9, 0xc0, 0x3a, 0xf8, 0xe6, 0xa0, 0xf6, 0x3b, 0x1d, 0xc4,
	0x3f, 0x6e, 0x6a, 0x0f, 0x75, 0xe4, 0xb2, 0x8c, 0x79, 0xe1, 0x39, 0x35, 0x5f, 0x23, 0xec, 0xa1,
	0xae, 0xac, 0xe2, 0x8a, 0x51, 0x55, 0xaf, 0x21, 0xbc, 0x86, 0xf8, 0x09, 0xea, 0xcd, 0x05, 0x24,
	0x54, 0x9b, 0x30, 0xa3, 0x3c, 0x0c, 0x6f, 0x09, 0x3c, 0x44, 0x5b, 0x29, 0xc8, 0x44, 0xd0, 0xb9,
	0x8e, 0xcd, 0xeb, 0x98, 0x6f, 0xef, 0x52, 0xf8, 0x19, 0x7a, 0x94, 0x17, 0x3c, 0x8e, 0x8a, 0x62,
	0x39, 0xcb, 0x04, 0x3f, 0x03, 0xe6, 0x75, 0x87, 0xf6, 0xc8, 0x0d, 0x77, 0xae, 0xe9, 0xa9, 0x61,
	0xff, 0x08, 0xd1, 0xdd, 0x38, 0xc4, 0xde, 0x3d, 0x85, 0x88, 0xfe, 0x7f, 0x88, 0xcf, 0x27, 0xa8,
	0xdb, 0x4c, 0x83, 0xb7, 0x50, 0xb7, 0xa4, 0x4c, 0x51, 0x96, 0xef, 0x5a, 0x1a, 0x68, 0x3f, 0x1a,
	0xd8, 0x78, 0x1b, 0xb9, 0x99, 0x00, 0x38, 0xd3, 0xa8, 0x85, 0x77, 0xd1, 0xf6, 0xc7, 0x13, 0xaa,
	0xa0, 0xa0, 0xd2, 0x34, 0x3b, 0x47, 0x6f, 0x2f, 0x56, 0xbe, 0x7d, 0xb9, 0xf2, 0xed, 0x9f, 0x2b,
	0xdf, 0xfe, 0xb2, 0xf6, 0xad, 0xcb, 0xb5, 0x6f, 0x7d, 0x5f, 0xfb, 0xd6, 0xfb, 0xc3, 0x3b, 0x26,
	0x27, 0x66, 0x91, 0x53, 0x5e, 0xb1, 0x34, 0xd2, 0xb1, 0x90, 0xe6, 0xb0, 0x3f, 0xdd, 0x9e, 0xb6,
	0x71, 0x1d, 0x77, 0xcc, 0x39, 0x1e, 0xfe, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x2a, 0xa0, 0x2e, 0x15,
	0xfa, 0x03, 0x00, 0x00,
}

func (m *Definition) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Definition) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Definition) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SendCommissionRate.Size()
		i -= size
		if _, err := m.SendCommissionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintToken(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	{
		size := m.BurnRate.Size()
		i -= size
		if _, err := m.BurnRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintToken(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
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
		i = encodeVarintToken(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Issuer) > 0 {
		i -= len(m.Issuer)
		copy(dAtA[i:], m.Issuer)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Issuer)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Token) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Token) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Token) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SendCommissionRate.Size()
		i -= size
		if _, err := m.SendCommissionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintToken(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x52
	{
		size := m.BurnRate.Size()
		i -= size
		if _, err := m.BurnRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintToken(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x4a
	if len(m.Features) > 0 {
		dAtA4 := make([]byte, len(m.Features)*10)
		var j3 int
		for _, num := range m.Features {
			for num >= 1<<7 {
				dAtA4[j3] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j3++
			}
			dAtA4[j3] = uint8(num)
			j3++
		}
		i -= j3
		copy(dAtA[i:], dAtA4[:j3])
		i = encodeVarintToken(dAtA, i, uint64(j3))
		i--
		dAtA[i] = 0x42
	}
	if m.GloballyFrozen {
		i--
		if m.GloballyFrozen {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x38
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x32
	}
	if m.Precision != 0 {
		i = encodeVarintToken(dAtA, i, uint64(m.Precision))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Subunit) > 0 {
		i -= len(m.Subunit)
		copy(dAtA[i:], m.Subunit)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Subunit)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Symbol) > 0 {
		i -= len(m.Symbol)
		copy(dAtA[i:], m.Symbol)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Symbol)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Issuer) > 0 {
		i -= len(m.Issuer)
		copy(dAtA[i:], m.Issuer)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Issuer)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintToken(dAtA []byte, offset int, v uint64) int {
	offset -= sovToken(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Definition) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	l = len(m.Issuer)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	if len(m.Features) > 0 {
		l = 0
		for _, e := range m.Features {
			l += sovToken(uint64(e))
		}
		n += 1 + sovToken(uint64(l)) + l
	}
	l = m.BurnRate.Size()
	n += 1 + l + sovToken(uint64(l))
	l = m.SendCommissionRate.Size()
	n += 1 + l + sovToken(uint64(l))
	return n
}

func (m *Token) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	l = len(m.Issuer)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	l = len(m.Symbol)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	l = len(m.Subunit)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	if m.Precision != 0 {
		n += 1 + sovToken(uint64(m.Precision))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	if m.GloballyFrozen {
		n += 2
	}
	if len(m.Features) > 0 {
		l = 0
		for _, e := range m.Features {
			l += sovToken(uint64(e))
		}
		n += 1 + sovToken(uint64(l)) + l
	}
	l = m.BurnRate.Size()
	n += 1 + l + sovToken(uint64(l))
	l = m.SendCommissionRate.Size()
	n += 1 + l + sovToken(uint64(l))
	return n
}

func sovToken(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozToken(x uint64) (n int) {
	return sovToken(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Definition) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowToken
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
			return fmt.Errorf("proto: Definition: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Definition: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
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
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Issuer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType == 0 {
				var v Feature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowToken
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= Feature(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Features = append(m.Features, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowToken
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
					return ErrInvalidLengthToken
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthToken
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Features) == 0 {
					m.Features = make([]Feature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v Feature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowToken
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= Feature(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Features = append(m.Features, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Features", wireType)
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.BurnRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SendCommissionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SendCommissionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipToken(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthToken
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
func (m *Token) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowToken
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
			return fmt.Errorf("proto: Token: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Token: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
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
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
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
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
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
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
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
					return ErrIntOverflowToken
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
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GloballyFrozen", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.GloballyFrozen = bool(v != 0)
		case 8:
			if wireType == 0 {
				var v Feature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowToken
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= Feature(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Features = append(m.Features, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowToken
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
					return ErrInvalidLengthToken
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthToken
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Features) == 0 {
					m.Features = make([]Feature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v Feature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowToken
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= Feature(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Features = append(m.Features, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Features", wireType)
			}
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.BurnRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SendCommissionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SendCommissionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipToken(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthToken
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
func skipToken(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowToken
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
					return 0, ErrIntOverflowToken
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
					return 0, ErrIntOverflowToken
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
				return 0, ErrInvalidLengthToken
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupToken
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthToken
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthToken        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowToken          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupToken = fmt.Errorf("proto: unexpected end of group")
)
