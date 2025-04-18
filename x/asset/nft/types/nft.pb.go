// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/nft/v1/nft.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/codec/types"
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

// ClassFeature defines possible features of non-fungible token class.
type ClassFeature int32

const (
	ClassFeature_burning         ClassFeature = 0
	ClassFeature_freezing        ClassFeature = 1
	ClassFeature_whitelisting    ClassFeature = 2
	ClassFeature_disable_sending ClassFeature = 3
	ClassFeature_soulbound       ClassFeature = 4
)

var ClassFeature_name = map[int32]string{
	0: "burning",
	1: "freezing",
	2: "whitelisting",
	3: "disable_sending",
	4: "soulbound",
}

var ClassFeature_value = map[string]int32{
	"burning":         0,
	"freezing":        1,
	"whitelisting":    2,
	"disable_sending": 3,
	"soulbound":       4,
}

func (x ClassFeature) String() string {
	return proto.EnumName(ClassFeature_name, int32(x))
}

func (ClassFeature) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5b9231d6a69d6d06, []int{0}
}

// ClassDefinition defines the non-fungible token class settings to store.
type ClassDefinition struct {
	ID       string         `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Issuer   string         `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Features []ClassFeature `protobuf:"varint,3,rep,packed,name=features,proto3,enum=coreum.asset.nft.v1.ClassFeature" json:"features,omitempty"`
	// royalty_rate is a number between 0 and 1,which will be used in coreum native DEX.
	// whenever an NFT this class is traded on the DEX, the traded amount will be multiplied by this value
	// that will be transferred to the issuer of the NFT.
	RoyaltyRate cosmossdk_io_math.LegacyDec `protobuf:"bytes,4,opt,name=royalty_rate,json=royaltyRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"royalty_rate"`
}

func (m *ClassDefinition) Reset()         { *m = ClassDefinition{} }
func (m *ClassDefinition) String() string { return proto.CompactTextString(m) }
func (*ClassDefinition) ProtoMessage()    {}
func (*ClassDefinition) Descriptor() ([]byte, []int) {
	return fileDescriptor_5b9231d6a69d6d06, []int{0}
}
func (m *ClassDefinition) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ClassDefinition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ClassDefinition.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ClassDefinition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClassDefinition.Merge(m, src)
}
func (m *ClassDefinition) XXX_Size() int {
	return m.Size()
}
func (m *ClassDefinition) XXX_DiscardUnknown() {
	xxx_messageInfo_ClassDefinition.DiscardUnknown(m)
}

var xxx_messageInfo_ClassDefinition proto.InternalMessageInfo

func (m *ClassDefinition) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *ClassDefinition) GetIssuer() string {
	if m != nil {
		return m.Issuer
	}
	return ""
}

func (m *ClassDefinition) GetFeatures() []ClassFeature {
	if m != nil {
		return m.Features
	}
	return nil
}

// Class is a full representation of the non-fungible token class.
type Class struct {
	Id          string         `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Issuer      string         `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Name        string         `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Symbol      string         `protobuf:"bytes,4,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Description string         `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	URI         string         `protobuf:"bytes,6,opt,name=uri,proto3" json:"uri,omitempty"`
	URIHash     string         `protobuf:"bytes,7,opt,name=uri_hash,json=uriHash,proto3" json:"uri_hash,omitempty"`
	Data        *types.Any     `protobuf:"bytes,8,opt,name=data,proto3" json:"data,omitempty"`
	Features    []ClassFeature `protobuf:"varint,9,rep,packed,name=features,proto3,enum=coreum.asset.nft.v1.ClassFeature" json:"features,omitempty"`
	// royalty_rate is a number between 0 and 1,which will be used in coreum native DEX.
	// whenever an NFT this class is traded on the DEX, the traded amount will be multiplied by this value
	// that will be transferred to the issuer of the NFT.
	RoyaltyRate cosmossdk_io_math.LegacyDec `protobuf:"bytes,10,opt,name=royalty_rate,json=royaltyRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"royalty_rate"`
}

func (m *Class) Reset()         { *m = Class{} }
func (m *Class) String() string { return proto.CompactTextString(m) }
func (*Class) ProtoMessage()    {}
func (*Class) Descriptor() ([]byte, []int) {
	return fileDescriptor_5b9231d6a69d6d06, []int{1}
}
func (m *Class) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Class) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Class.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Class) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Class.Merge(m, src)
}
func (m *Class) XXX_Size() int {
	return m.Size()
}
func (m *Class) XXX_DiscardUnknown() {
	xxx_messageInfo_Class.DiscardUnknown(m)
}

var xxx_messageInfo_Class proto.InternalMessageInfo

func (m *Class) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Class) GetIssuer() string {
	if m != nil {
		return m.Issuer
	}
	return ""
}

func (m *Class) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Class) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *Class) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Class) GetURI() string {
	if m != nil {
		return m.URI
	}
	return ""
}

func (m *Class) GetURIHash() string {
	if m != nil {
		return m.URIHash
	}
	return ""
}

func (m *Class) GetData() *types.Any {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Class) GetFeatures() []ClassFeature {
	if m != nil {
		return m.Features
	}
	return nil
}

func init() {
	proto.RegisterEnum("coreum.asset.nft.v1.ClassFeature", ClassFeature_name, ClassFeature_value)
	proto.RegisterType((*ClassDefinition)(nil), "coreum.asset.nft.v1.ClassDefinition")
	proto.RegisterType((*Class)(nil), "coreum.asset.nft.v1.Class")
}

func init() { proto.RegisterFile("coreum/asset/nft/v1/nft.proto", fileDescriptor_5b9231d6a69d6d06) }

var fileDescriptor_5b9231d6a69d6d06 = []byte{
	// 520 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x53, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x8e, 0xed, 0x34, 0x3f, 0x9b, 0xd0, 0x46, 0xdb, 0xaa, 0x72, 0x8b, 0x70, 0x42, 0x91, 0x50,
	0xc4, 0xc1, 0x56, 0x8b, 0x04, 0x27, 0x0e, 0xb4, 0x51, 0x44, 0x24, 0x2e, 0xac, 0xd4, 0x0b, 0x97,
	0x68, 0x6d, 0x6f, 0xec, 0x15, 0xb6, 0x37, 0xda, 0x9f, 0x80, 0x79, 0x0a, 0x1e, 0x2b, 0xc7, 0x1e,
	0x11, 0x87, 0x08, 0x39, 0x4f, 0xc0, 0x1b, 0xa0, 0x5d, 0x07, 0x08, 0x12, 0xe2, 0x00, 0xa7, 0x9d,
	0xf9, 0xbe, 0x19, 0xcd, 0xb7, 0xdf, 0x68, 0xc0, 0x83, 0x88, 0x71, 0xa2, 0xf2, 0x00, 0x0b, 0x41,
	0x64, 0x50, 0x2c, 0x64, 0xb0, 0xba, 0xd4, 0x8f, 0xbf, 0xe4, 0x4c, 0x32, 0x78, 0x5c, 0xd3, 0xbe,
	0xa1, 0x7d, 0x8d, 0xaf, 0x2e, 0xcf, 0x4f, 0x12, 0x96, 0x30, 0xc3, 0x07, 0x3a, 0xaa, 0x4b, 0xcf,
	0xcf, 0x12, 0xc6, 0x92, 0x8c, 0x04, 0x26, 0x0b, 0xd5, 0x22, 0xc0, 0x45, 0x59, 0x53, 0x17, 0x6b,
	0x0b, 0x1c, 0xdd, 0x64, 0x58, 0x88, 0x09, 0x59, 0xd0, 0x82, 0x4a, 0xca, 0x0a, 0x78, 0x0a, 0x6c,
	0x1a, 0xbb, 0xd6, 0xc8, 0x1a, 0x77, 0xaf, 0x5b, 0xd5, 0x66, 0x68, 0xcf, 0x26, 0xc8, 0xa6, 0x31,
	0x3c, 0x05, 0x2d, 0x2a, 0x84, 0x22, 0xdc, 0xb5, 0x35, 0x87, 0x76, 0x19, 0x7c, 0x01, 0x3a, 0x0b,
	0x82, 0xa5, 0xe2, 0x44, 0xb8, 0xce, 0xc8, 0x19, 0x1f, 0x5e, 0x3d, 0xf4, 0xff, 0x20, 0xce, 0x37,
	0x73, 0xa6, 0x75, 0x25, 0xfa, 0xd9, 0x02, 0xa7, 0xa0, 0xcf, 0x59, 0x89, 0x33, 0x59, 0xce, 0x39,
	0x96, 0xc4, 0x6d, 0x9a, 0xc1, 0x8f, 0xd6, 0x9b, 0x61, 0xe3, 0xcb, 0x66, 0x78, 0x3f, 0x62, 0x22,
	0x67, 0x42, 0xc4, 0xef, 0x7c, 0xca, 0x82, 0x1c, 0xcb, 0xd4, 0x7f, 0x4d, 0x12, 0x1c, 0x95, 0x13,
	0x12, 0xa1, 0xde, 0xae, 0x11, 0x61, 0x49, 0x2e, 0xbe, 0xd9, 0xe0, 0xc0, 0x8c, 0x80, 0x87, 0xbf,
	0x3e, 0xf0, 0x57, 0xe1, 0x10, 0x34, 0x0b, 0x9c, 0x13, 0xd7, 0x31, 0xa8, 0x89, 0x75, 0xad, 0x28,
	0xf3, 0x90, 0x65, 0xb5, 0x0e, 0xb4, 0xcb, 0xe0, 0x08, 0xf4, 0x62, 0x22, 0x22, 0x4e, 0x97, 0xda,
	0x23, 0xf7, 0xc0, 0x90, 0xfb, 0x10, 0x3c, 0x03, 0x8e, 0xe2, 0xd4, 0x6d, 0x19, 0xf9, 0xed, 0x6a,
	0x33, 0x74, 0x6e, 0xd1, 0x0c, 0x69, 0x0c, 0x3e, 0x06, 0x1d, 0xc5, 0xe9, 0x3c, 0xc5, 0x22, 0x75,
	0xdb, 0x86, 0xef, 0x55, 0x9b, 0x61, 0xfb, 0x16, 0xcd, 0x5e, 0x61, 0x91, 0xa2, 0xb6, 0xe2, 0x54,
	0x07, 0x70, 0x0c, 0x9a, 0x31, 0x96, 0xd8, 0xed, 0x8c, 0xac, 0x71, 0xef, 0xea, 0xc4, 0xaf, 0xf7,
	0xe6, 0xff, 0xd8, 0x9b, 0xff, 0xb2, 0x28, 0x91, 0xa9, 0xf8, 0xcd, 0xf3, 0xee, 0xff, 0x7b, 0x0e,
	0xfe, 0xcd, 0xf3, 0x27, 0x73, 0xd0, 0xdf, 0x9f, 0x00, 0x7b, 0xa0, 0x1d, 0x2a, 0x5e, 0xd0, 0x22,
	0x19, 0x34, 0x60, 0x1f, 0x74, 0x16, 0x9c, 0x90, 0x8f, 0x3a, 0xb3, 0xe0, 0x00, 0xf4, 0xdf, 0xa7,
	0x54, 0x92, 0x8c, 0x0a, 0xa9, 0x11, 0x1b, 0x1e, 0x83, 0xa3, 0x98, 0x0a, 0x1c, 0x66, 0x64, 0x2e,
	0x48, 0x11, 0x6b, 0xd0, 0x81, 0xf7, 0x40, 0x57, 0x30, 0x95, 0x85, 0x4c, 0x15, 0xf1, 0xa0, 0x79,
	0xfd, 0x66, 0x5d, 0x79, 0xd6, 0x5d, 0xe5, 0x59, 0x5f, 0x2b, 0xcf, 0xfa, 0xb4, 0xf5, 0x1a, 0x77,
	0x5b, 0xaf, 0xf1, 0x79, 0xeb, 0x35, 0xde, 0x3e, 0x4f, 0xa8, 0x4c, 0x55, 0xe8, 0x47, 0x2c, 0x0f,
	0x6e, 0xcc, 0xcf, 0xa7, 0xba, 0x07, 0xeb, 0x5d, 0x04, 0xbb, 0xd3, 0x59, 0x3d, 0x0b, 0x3e, 0xec,
	0xdd, 0x8f, 0x2c, 0x97, 0x44, 0x84, 0x2d, 0x63, 0xe7, 0xd3, 0xef, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x3e, 0xcc, 0xd0, 0x0d, 0x60, 0x03, 0x00, 0x00,
}

func (m *ClassDefinition) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ClassDefinition) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ClassDefinition) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.RoyaltyRate.Size()
		i -= size
		if _, err := m.RoyaltyRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintNft(dAtA, i, uint64(size))
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
		i = encodeVarintNft(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Issuer) > 0 {
		i -= len(m.Issuer)
		copy(dAtA[i:], m.Issuer)
		i = encodeVarintNft(dAtA, i, uint64(len(m.Issuer)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ID) > 0 {
		i -= len(m.ID)
		copy(dAtA[i:], m.ID)
		i = encodeVarintNft(dAtA, i, uint64(len(m.ID)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Class) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Class) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Class) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.RoyaltyRate.Size()
		i -= size
		if _, err := m.RoyaltyRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintNft(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x52
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
		i = encodeVarintNft(dAtA, i, uint64(j3))
		i--
		dAtA[i] = 0x4a
	}
	if m.Data != nil {
		{
			size, err := m.Data.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintNft(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x42
	}
	if len(m.URIHash) > 0 {
		i -= len(m.URIHash)
		copy(dAtA[i:], m.URIHash)
		i = encodeVarintNft(dAtA, i, uint64(len(m.URIHash)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.URI) > 0 {
		i -= len(m.URI)
		copy(dAtA[i:], m.URI)
		i = encodeVarintNft(dAtA, i, uint64(len(m.URI)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintNft(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Symbol) > 0 {
		i -= len(m.Symbol)
		copy(dAtA[i:], m.Symbol)
		i = encodeVarintNft(dAtA, i, uint64(len(m.Symbol)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintNft(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Issuer) > 0 {
		i -= len(m.Issuer)
		copy(dAtA[i:], m.Issuer)
		i = encodeVarintNft(dAtA, i, uint64(len(m.Issuer)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintNft(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintNft(dAtA []byte, offset int, v uint64) int {
	offset -= sovNft(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ClassDefinition) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.Issuer)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	if len(m.Features) > 0 {
		l = 0
		for _, e := range m.Features {
			l += sovNft(uint64(e))
		}
		n += 1 + sovNft(uint64(l)) + l
	}
	l = m.RoyaltyRate.Size()
	n += 1 + l + sovNft(uint64(l))
	return n
}

func (m *Class) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.Issuer)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.Symbol)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.URI)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	l = len(m.URIHash)
	if l > 0 {
		n += 1 + l + sovNft(uint64(l))
	}
	if m.Data != nil {
		l = m.Data.Size()
		n += 1 + l + sovNft(uint64(l))
	}
	if len(m.Features) > 0 {
		l = 0
		for _, e := range m.Features {
			l += sovNft(uint64(e))
		}
		n += 1 + sovNft(uint64(l)) + l
	}
	l = m.RoyaltyRate.Size()
	n += 1 + l + sovNft(uint64(l))
	return n
}

func sovNft(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozNft(x uint64) (n int) {
	return sovNft(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ClassDefinition) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNft
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
			return fmt.Errorf("proto: ClassDefinition: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ClassDefinition: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Issuer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Issuer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType == 0 {
				var v ClassFeature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowNft
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= ClassFeature(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Features = append(m.Features, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowNft
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
					return ErrInvalidLengthNft
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthNft
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Features) == 0 {
					m.Features = make([]ClassFeature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v ClassFeature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowNft
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= ClassFeature(b&0x7F) << shift
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
				return fmt.Errorf("proto: wrong wireType = %d for field RoyaltyRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RoyaltyRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthNft
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
func (m *Class) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNft
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
			return fmt.Errorf("proto: Class: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Class: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Issuer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Issuer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Symbol", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Symbol = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field URI", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.URI = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field URIHash", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.URIHash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Data == nil {
				m.Data = &types.Any{}
			}
			if err := m.Data.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType == 0 {
				var v ClassFeature
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowNft
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= ClassFeature(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Features = append(m.Features, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowNft
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
					return ErrInvalidLengthNft
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthNft
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Features) == 0 {
					m.Features = make([]ClassFeature, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v ClassFeature
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowNft
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= ClassFeature(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Features = append(m.Features, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Features", wireType)
			}
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoyaltyRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNft
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
				return ErrInvalidLengthNft
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNft
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RoyaltyRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthNft
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
func skipNft(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowNft
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
					return 0, ErrIntOverflowNft
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
					return 0, ErrIntOverflowNft
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
				return 0, ErrInvalidLengthNft
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupNft
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthNft
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthNft        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowNft          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupNft = fmt.Errorf("proto: unexpected end of group")
)
