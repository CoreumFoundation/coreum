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

// EventNonFungibleTokenClassIssued is emitted on MsgIssueNonFungibleTokenClass.
type EventNonFungibleTokenClassIssued struct {
	ID          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Issuer      string `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	Symbol      string `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Name        string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	URI         string `protobuf:"bytes,6,opt,name=uri,proto3" json:"uri,omitempty"`
	URIHash     string `protobuf:"bytes,7,opt,name=uri_hash,json=uriHash,proto3" json:"uri_hash,omitempty"`
}

func (m *EventNonFungibleTokenClassIssued) Reset()         { *m = EventNonFungibleTokenClassIssued{} }
func (m *EventNonFungibleTokenClassIssued) String() string { return proto.CompactTextString(m) }
func (*EventNonFungibleTokenClassIssued) ProtoMessage()    {}
func (*EventNonFungibleTokenClassIssued) Descriptor() ([]byte, []int) {
	return fileDescriptor_aede4b64fdc52aa3, []int{2}
}
func (m *EventNonFungibleTokenClassIssued) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventNonFungibleTokenClassIssued) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventNonFungibleTokenClassIssued.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventNonFungibleTokenClassIssued) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventNonFungibleTokenClassIssued.Merge(m, src)
}
func (m *EventNonFungibleTokenClassIssued) XXX_Size() int {
	return m.Size()
}
func (m *EventNonFungibleTokenClassIssued) XXX_DiscardUnknown() {
	xxx_messageInfo_EventNonFungibleTokenClassIssued.DiscardUnknown(m)
}

var xxx_messageInfo_EventNonFungibleTokenClassIssued proto.InternalMessageInfo

func (m *EventNonFungibleTokenClassIssued) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *EventNonFungibleTokenClassIssued) GetIssuer() string {
	if m != nil {
		return m.Issuer
	}
	return ""
}

func (m *EventNonFungibleTokenClassIssued) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *EventNonFungibleTokenClassIssued) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *EventNonFungibleTokenClassIssued) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *EventNonFungibleTokenClassIssued) GetURI() string {
	if m != nil {
		return m.URI
	}
	return ""
}

func (m *EventNonFungibleTokenClassIssued) GetURIHash() string {
	if m != nil {
		return m.URIHash
	}
	return ""
}

func init() {
	proto.RegisterType((*EventFungibleTokenIssued)(nil), "coreum.asset.v1.EventFungibleTokenIssued")
	proto.RegisterType((*EventFungibleTokenFrozenAmountChanged)(nil), "coreum.asset.v1.EventFungibleTokenFrozenAmountChanged")
	proto.RegisterType((*EventNonFungibleTokenClassIssued)(nil), "coreum.asset.v1.EventNonFungibleTokenClassIssued")
}

func init() { proto.RegisterFile("coreum/asset/v1/event.proto", fileDescriptor_aede4b64fdc52aa3) }

var fileDescriptor_aede4b64fdc52aa3 = []byte{
	// 575 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0xcd, 0x6e, 0xda, 0x40,
	0x10, 0xc6, 0x90, 0x40, 0x58, 0x04, 0x91, 0xac, 0x28, 0x72, 0xd2, 0xca, 0x20, 0xa4, 0x44, 0x5c,
	0xba, 0x2e, 0xe9, 0x13, 0x04, 0x5a, 0x14, 0x7a, 0xe8, 0xc1, 0x2a, 0x97, 0x5e, 0xa2, 0xb5, 0xbd,
	0x85, 0x55, 0xf0, 0xae, 0xb5, 0x3f, 0xa8, 0xe9, 0x53, 0xf4, 0xb1, 0x72, 0x8c, 0xd4, 0x4b, 0xd5,
	0x03, 0xaa, 0xcc, 0x1b, 0xf4, 0x09, 0xaa, 0xfd, 0xa1, 0xf9, 0xe1, 0xd2, 0x9e, 0x3c, 0xf3, 0x7d,
	0x3b, 0x33, 0x9e, 0x99, 0x6f, 0xc0, 0x8b, 0x94, 0x71, 0xac, 0xf2, 0x08, 0x09, 0x81, 0x65, 0xb4,
	0x1a, 0x46, 0x78, 0x85, 0xa9, 0x84, 0x05, 0x67, 0x92, 0xf9, 0x87, 0x96, 0x84, 0x86, 0x84, 0xab,
	0xe1, 0xe9, 0xd1, 0x9c, 0xcd, 0x99, 0xe1, 0x22, 0x6d, 0xd9, 0x67, 0xa7, 0x61, 0xca, 0x44, 0xce,
	0x44, 0x94, 0x20, 0x81, 0xa3, 0xd5, 0x30, 0xc1, 0x12, 0x0d, 0xa3, 0x94, 0x11, 0xea, 0xf8, 0x9d,
	0x1a, 0x36, 0x9f, 0x21, 0xfb, 0xbf, 0xab, 0x20, 0x78, 0xa7, 0x6b, 0x4e, 0x14, 0x9d, 0x93, 0x64,
	0x89, 0x3f, 0xb2, 0x1b, 0x4c, 0xa7, 0x42, 0x28, 0x9c, 0xf9, 0x47, 0x60, 0x3f, 0xc3, 0x94, 0xe5,
	0x81, 0xd7, 0xf3, 0x06, 0xcd, 0xd8, 0x3a, 0xfe, 0x31, 0xa8, 0x13, 0xcd, 0xf3, 0xa0, 0x6a, 0x60,
	0xe7, 0x69, 0x5c, 0xdc, 0xe6, 0x09, 0x5b, 0x06, 0x35, 0x8b, 0x5b, 0xcf, 0x0f, 0x40, 0x43, 0xa8,
	0x44, 0x51, 0x22, 0x83, 0x3d, 0x43, 0x6c, 0x5d, 0xff, 0x25, 0x68, 0x16, 0x1c, 0xa7, 0x44, 0x10,
	0x46, 0x83, 0xfd, 0x9e, 0x37, 0x68, 0xc7, 0x0f, 0x80, 0x66, 0xb5, 0x5d, 0x10, 0x4c, 0x65, 0x50,
	0x37, 0x91, 0x0f, 0x80, 0x3f, 0x03, 0x1d, 0x42, 0x89, 0x24, 0x68, 0x79, 0x8d, 0x72, 0xa6, 0xa8,
	0x0c, 0x1a, 0xfa, 0xc9, 0x08, 0xde, 0xad, 0xbb, 0x95, 0x9f, 0xeb, 0xee, 0xf9, 0x9c, 0xc8, 0x85,
	0x4a, 0x60, 0xca, 0xf2, 0xc8, 0x0d, 0xc8, 0x7e, 0x5e, 0x89, 0xec, 0x26, 0x92, 0xb7, 0x05, 0x16,
	0x70, 0x4a, 0x65, 0xdc, 0x76, 0x59, 0x2e, 0x4d, 0x12, 0xbf, 0x07, 0x5a, 0x19, 0x16, 0x29, 0x27,
	0x85, 0xd4, 0x3f, 0x75, 0x60, 0xca, 0x3e, 0x86, 0xfc, 0x4b, 0x70, 0xf0, 0x19, 0x23, 0xa9, 0x38,
	0x16, 0x41, 0xb3, 0x57, 0x1b, 0x74, 0x2e, 0xce, 0xe0, 0xb3, 0x45, 0xc1, 0x27, 0xc3, 0x9c, 0xd8,
	0xd7, 0xf1, 0xdf, 0xb0, 0xfe, 0x77, 0x0f, 0x9c, 0xed, 0x0e, 0x7d, 0xc2, 0xd9, 0x57, 0x4c, 0xed,
	0x7f, 0x8c, 0x17, 0x88, 0xce, 0x71, 0xa6, 0x67, 0x87, 0xd2, 0xd4, 0xb4, 0x67, 0x77, 0xb0, 0x75,
	0xfd, 0x2b, 0x70, 0x58, 0x70, 0xbc, 0x22, 0x4c, 0x89, 0xed, 0x00, 0xf4, 0x3a, 0x5a, 0x17, 0x27,
	0xd0, 0xf6, 0x09, 0xb5, 0x1e, 0xa0, 0xd3, 0x03, 0x1c, 0x33, 0x42, 0x47, 0x7b, 0x7a, 0x36, 0x71,
	0x67, 0x1b, 0xe7, 0x5a, 0x9e, 0x80, 0x4e, 0xaa, 0x38, 0xc7, 0x54, 0x6e, 0x13, 0xd5, 0xfe, 0x2d,
	0x51, 0xdb, 0x85, 0xd9, 0x3c, 0xfd, 0x8d, 0x07, 0x7a, 0xa6, 0xab, 0x0f, 0x8c, 0x3e, 0x69, 0x6c,
	0xbc, 0x44, 0x42, 0x38, 0x49, 0x1d, 0x83, 0x2a, 0xc9, 0x6c, 0x2f, 0xa3, 0x7a, 0xb9, 0xee, 0x56,
	0xa7, 0x6f, 0xe3, 0x2a, 0xc9, 0xfe, 0x5b, 0x54, 0x3e, 0xd8, 0xa3, 0x28, 0xc7, 0x4e, 0x51, 0xc6,
	0x7e, 0xbe, 0xbb, 0xfd, 0xdd, 0xdd, 0x9d, 0x80, 0x9a, 0xe2, 0xc4, 0x8a, 0x69, 0xd4, 0x28, 0xd7,
	0xdd, 0xda, 0x2c, 0x9e, 0xc6, 0x1a, 0xf3, 0xcf, 0xc1, 0x81, 0xe2, 0xe4, 0x7a, 0x81, 0xc4, 0xc2,
	0x29, 0xa9, 0x55, 0xae, 0xbb, 0x8d, 0x59, 0x3c, 0xbd, 0x42, 0x62, 0x11, 0x37, 0x14, 0x27, 0xda,
	0x18, 0xbd, 0xbf, 0x2b, 0x43, 0xef, 0xbe, 0x0c, 0xbd, 0x5f, 0x65, 0xe8, 0x7d, 0xdb, 0x84, 0x95,
	0xfb, 0x4d, 0x58, 0xf9, 0xb1, 0x09, 0x2b, 0x9f, 0x5e, 0x3f, 0x52, 0xdc, 0xd8, 0x08, 0x62, 0xc2,
	0x14, 0xcd, 0x90, 0xae, 0x1c, 0xb9, 0x1b, 0xfc, 0xe2, 0xae, 0xd0, 0xe8, 0x2f, 0xa9, 0x9b, 0x1b,
	0x7c, 0xf3, 0x27, 0x00, 0x00, 0xff, 0xff, 0x73, 0xa8, 0xde, 0x69, 0x06, 0x04, 0x00, 0x00,
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

func (m *EventNonFungibleTokenClassIssued) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventNonFungibleTokenClassIssued) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventNonFungibleTokenClassIssued) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.URIHash) > 0 {
		i -= len(m.URIHash)
		copy(dAtA[i:], m.URIHash)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.URIHash)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.URI) > 0 {
		i -= len(m.URI)
		copy(dAtA[i:], m.URI)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.URI)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Name)))
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
	if len(m.ID) > 0 {
		i -= len(m.ID)
		copy(dAtA[i:], m.ID)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.ID)))
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

func (m *EventNonFungibleTokenClassIssued) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ID)
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
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.URI)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.URIHash)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
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
func (m *EventNonFungibleTokenClassIssued) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: EventNonFungibleTokenClassIssued: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventNonFungibleTokenClassIssued: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
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
			m.ID = string(dAtA[iNdEx:postIndex])
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
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
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
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
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
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field URI", wireType)
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
			m.URI = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field URIHash", wireType)
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
			m.URIHash = string(dAtA[iNdEx:postIndex])
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
