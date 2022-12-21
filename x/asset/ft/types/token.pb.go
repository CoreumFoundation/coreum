// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/ft/v1/token.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
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

// TokenFeature defines possible features of fungible token
type TokenFeature int32

const (
	TokenFeature_freeze    TokenFeature = 0
	TokenFeature_mint      TokenFeature = 1
	TokenFeature_burn      TokenFeature = 2
	TokenFeature_whitelist TokenFeature = 3
)

var TokenFeature_name = map[int32]string{
	0: "freeze",
	1: "mint",
	2: "burn",
	3: "whitelist",
}

var TokenFeature_value = map[string]int32{
	"freeze":    0,
	"mint":      1,
	"burn":      2,
	"whitelist": 3,
}

func (x TokenFeature) String() string {
	return proto.EnumName(TokenFeature_name, int32(x))
}

func (TokenFeature) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fe80c7a2c55589e7, []int{0}
}

// FTDefinition defines the fungible token settings to store.
type FTDefinition struct {
	Denom    string         `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Issuer   string         `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Features []TokenFeature `protobuf:"varint,3,rep,packed,name=features,proto3,enum=coreum.asset.ft.v1.TokenFeature" json:"features,omitempty"`
	// burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine
	// burn_amount. This value will be burnt on top of the send amount.
	BurnRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,4,opt,name=burn_rate,json=burnRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"burn_rate"`
}

func (m *FTDefinition) Reset()         { *m = FTDefinition{} }
func (m *FTDefinition) String() string { return proto.CompactTextString(m) }
func (*FTDefinition) ProtoMessage()    {}
func (*FTDefinition) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe80c7a2c55589e7, []int{0}
}
func (m *FTDefinition) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FTDefinition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FTDefinition.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FTDefinition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FTDefinition.Merge(m, src)
}
func (m *FTDefinition) XXX_Size() int {
	return m.Size()
}
func (m *FTDefinition) XXX_DiscardUnknown() {
	xxx_messageInfo_FTDefinition.DiscardUnknown(m)
}

var xxx_messageInfo_FTDefinition proto.InternalMessageInfo

// FT is a full representation of the fungible token.
type FT struct {
	Denom          string         `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Issuer         string         `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Symbol         string         `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Subunit        string         `protobuf:"bytes,4,opt,name=subunit,proto3" json:"subunit,omitempty"`
	Precision      uint32         `protobuf:"varint,5,opt,name=precision,proto3" json:"precision,omitempty"`
	Description    string         `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	GloballyFrozen bool           `protobuf:"varint,7,opt,name=globally_frozen,json=globallyFrozen,proto3" json:"globally_frozen,omitempty"`
	Features       []TokenFeature `protobuf:"varint,8,rep,packed,name=features,proto3,enum=coreum.asset.ft.v1.TokenFeature" json:"features,omitempty"`
	// burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine
	// burn_amount. This value will be burnt on top of the send amount.
	BurnRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,9,opt,name=burn_rate,json=burnRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"burn_rate"`
}

func (m *FT) Reset()         { *m = FT{} }
func (m *FT) String() string { return proto.CompactTextString(m) }
func (*FT) ProtoMessage()    {}
func (*FT) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe80c7a2c55589e7, []int{1}
}
func (m *FT) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FT) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FT.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FT) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FT.Merge(m, src)
}
func (m *FT) XXX_Size() int {
	return m.Size()
}
func (m *FT) XXX_DiscardUnknown() {
	xxx_messageInfo_FT.DiscardUnknown(m)
}

var xxx_messageInfo_FT proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("coreum.asset.ft.v1.TokenFeature", TokenFeature_name, TokenFeature_value)
	proto.RegisterType((*FTDefinition)(nil), "coreum.asset.ft.v1.FTDefinition")
	proto.RegisterType((*FT)(nil), "coreum.asset.ft.v1.FT")
}

func init() { proto.RegisterFile("coreum/asset/ft/v1/token.proto", fileDescriptor_fe80c7a2c55589e7) }

var fileDescriptor_fe80c7a2c55589e7 = []byte{
	// 472 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0x3f, 0x6f, 0xd3, 0x40,
	0x14, 0xf7, 0x25, 0xad, 0x6b, 0x1f, 0x6d, 0x89, 0x4e, 0x55, 0x75, 0xaa, 0x90, 0x63, 0x75, 0x80,
	0x08, 0x89, 0x3b, 0x85, 0x6e, 0x08, 0x96, 0x52, 0x79, 0x41, 0x2c, 0x56, 0x26, 0x96, 0xca, 0x76,
	0x9e, 0xd3, 0x53, 0xed, 0xbb, 0xe8, 0xee, 0x1c, 0x48, 0x3f, 0x01, 0x23, 0x1f, 0xa1, 0x1f, 0xa7,
	0x63, 0xd9, 0x10, 0x43, 0x85, 0x92, 0x85, 0x8f, 0x81, 0x6c, 0xa7, 0x24, 0x88, 0x09, 0xd1, 0xc9,
	0xef, 0xf7, 0x7b, 0x7f, 0xfc, 0x7b, 0x3f, 0xdd, 0xc3, 0x41, 0xa6, 0x34, 0x54, 0x25, 0x4f, 0x8c,
	0x01, 0xcb, 0x73, 0xcb, 0x67, 0x43, 0x6e, 0xd5, 0x25, 0x48, 0x36, 0xd5, 0xca, 0x2a, 0x42, 0xda,
	0x3c, 0x6b, 0xf2, 0x2c, 0xb7, 0x6c, 0x36, 0x3c, 0x3a, 0x98, 0xa8, 0x89, 0x6a, 0xd2, 0xbc, 0x8e,
	0xda, 0xca, 0xa3, 0x20, 0x53, 0xa6, 0x54, 0x86, 0xa7, 0x89, 0x01, 0x3e, 0x1b, 0xa6, 0x60, 0x93,
	0x21, 0xcf, 0x94, 0x58, 0x4d, 0x3a, 0xfe, 0x8a, 0xf0, 0x6e, 0x34, 0x3a, 0x83, 0x5c, 0x48, 0x61,
	0x85, 0x92, 0xe4, 0x00, 0x6f, 0x8f, 0x41, 0xaa, 0x92, 0xa2, 0x10, 0x0d, 0xfc, 0xb8, 0x05, 0xe4,
	0x10, 0xbb, 0xc2, 0x98, 0x0a, 0x34, 0xed, 0x34, 0xf4, 0x0a, 0x91, 0xd7, 0xd8, 0xcb, 0x21, 0xb1,
	0x95, 0x06, 0x43, 0xbb, 0x61, 0x77, 0xb0, 0xff, 0x32, 0x64, 0x7f, 0x6b, 0x63, 0xa3, 0x5a, 0x7b,
	0xd4, 0x16, 0xc6, 0xbf, 0x3b, 0xc8, 0x3b, 0xec, 0xa7, 0x95, 0x96, 0xe7, 0x3a, 0xb1, 0x40, 0xb7,
	0xea, 0xc1, 0xa7, 0xec, 0xe6, 0xae, 0xef, 0x7c, 0xbf, 0xeb, 0x3f, 0x9d, 0x08, 0x7b, 0x51, 0xa5,
	0x2c, 0x53, 0x25, 0x5f, 0xad, 0xd0, 0x7e, 0x5e, 0x98, 0xf1, 0x25, 0xb7, 0xf3, 0x29, 0x18, 0x76,
	0x06, 0x59, 0xec, 0xd5, 0x03, 0xe2, 0xc4, 0xc2, 0x2b, 0xef, 0xf3, 0x75, 0xdf, 0xf9, 0x79, 0xdd,
	0x77, 0x8e, 0x17, 0x1d, 0xdc, 0x89, 0x46, 0xff, 0xb8, 0xc9, 0x21, 0x76, 0xcd, 0xbc, 0x4c, 0x55,
	0x41, 0xbb, 0x2d, 0xdf, 0x22, 0x42, 0xf1, 0x8e, 0xa9, 0xd2, 0x4a, 0x0a, 0xdb, 0x2a, 0x8c, 0xef,
	0x21, 0x79, 0x82, 0xfd, 0xa9, 0x86, 0x4c, 0x18, 0xa1, 0x24, 0xdd, 0x0e, 0xd1, 0x60, 0x2f, 0x5e,
	0x13, 0x24, 0xc4, 0x8f, 0xc6, 0x60, 0x32, 0x2d, 0xa6, 0xb5, 0xad, 0xd4, 0x6d, 0x7a, 0x37, 0x29,
	0xf2, 0x0c, 0x3f, 0x9e, 0x14, 0x2a, 0x4d, 0x8a, 0x62, 0x7e, 0x9e, 0x6b, 0x75, 0x05, 0x92, 0xee,
	0x84, 0x68, 0xe0, 0xc5, 0xfb, 0xf7, 0x74, 0xd4, 0xb0, 0x7f, 0x98, 0xec, 0xfd, 0x9f, 0xc9, 0xfe,
	0x43, 0x99, 0xfc, 0xfc, 0x0d, 0xde, 0xdd, 0xfc, 0x21, 0xc1, 0xd8, 0xcd, 0x35, 0xc0, 0x15, 0xf4,
	0x1c, 0xe2, 0xe1, 0xad, 0x52, 0x48, 0xdb, 0x43, 0x75, 0x54, 0xf7, 0xf6, 0x3a, 0x64, 0x0f, 0xfb,
	0x1f, 0x2f, 0x84, 0x85, 0x42, 0x18, 0xdb, 0xeb, 0x9e, 0xbe, 0xbf, 0x59, 0x04, 0xe8, 0x76, 0x11,
	0xa0, 0x1f, 0x8b, 0x00, 0x7d, 0x59, 0x06, 0xce, 0xed, 0x32, 0x70, 0xbe, 0x2d, 0x03, 0xe7, 0xc3,
	0xc9, 0x86, 0xa8, 0xb7, 0xcd, 0x96, 0x91, 0xaa, 0xe4, 0x38, 0xa9, 0x3d, 0xe3, 0xab, 0xbb, 0xf8,
	0xb4, 0xbe, 0x8c, 0x46, 0x65, 0xea, 0x36, 0xaf, 0xf9, 0xe4, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xa4, 0x3d, 0x09, 0x5a, 0x39, 0x03, 0x00, 0x00,
}

func (m *FTDefinition) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FTDefinition) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FTDefinition) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
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

func (m *FT) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FT) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FT) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
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
func (m *FTDefinition) Size() (n int) {
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
	return n
}

func (m *FT) Size() (n int) {
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
	return n
}

func sovToken(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozToken(x uint64) (n int) {
	return sovToken(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FTDefinition) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: FTDefinition: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FTDefinition: illegal tag %d (wire type %d)", fieldNum, wire)
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
				var v TokenFeature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowToken
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= TokenFeature(b&0x7F) << shift
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
					m.Features = make([]TokenFeature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v TokenFeature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowToken
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= TokenFeature(b&0x7F) << shift
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
func (m *FT) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: FT: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FT: illegal tag %d (wire type %d)", fieldNum, wire)
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
				var v TokenFeature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowToken
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= TokenFeature(b&0x7F) << shift
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
					m.Features = make([]TokenFeature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v TokenFeature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowToken
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= TokenFeature(b&0x7F) << shift
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
