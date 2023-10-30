// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/ft/v1/genesis.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
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
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// tokens keep the fungible token state
	Tokens []Token `protobuf:"bytes,2,rep,name=tokens,proto3" json:"tokens"`
	// frozen_balances contains the frozen balances on all of the accounts
	FrozenBalances []Balance `protobuf:"bytes,3,rep,name=frozen_balances,json=frozenBalances,proto3" json:"frozen_balances"`
	// whitelisted_balances contains the whitelisted balances on all of the accounts
	WhitelistedBalances []Balance `protobuf:"bytes,4,rep,name=whitelisted_balances,json=whitelistedBalances,proto3" json:"whitelisted_balances"`
	// pending_token_upgrades contains pending token upgrades.
	PendingTokenUpgrades []PendingTokenUpgrade `protobuf:"bytes,5,rep,name=pending_token_upgrades,json=pendingTokenUpgrades,proto3" json:"pending_token_upgrades"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_d281657d6c91cb92, []int{0}
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

func (m *GenesisState) GetTokens() []Token {
	if m != nil {
		return m.Tokens
	}
	return nil
}

func (m *GenesisState) GetFrozenBalances() []Balance {
	if m != nil {
		return m.FrozenBalances
	}
	return nil
}

func (m *GenesisState) GetWhitelistedBalances() []Balance {
	if m != nil {
		return m.WhitelistedBalances
	}
	return nil
}

func (m *GenesisState) GetPendingTokenUpgrades() []PendingTokenUpgrade {
	if m != nil {
		return m.PendingTokenUpgrades
	}
	return nil
}

// Balance defines an account address and balance pair used module genesis genesis state.
type Balance struct {
	// address is the address of the balance holder.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// coins defines the different coins this balance holds.
	Coins github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=coins,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"coins"`
}

func (m *Balance) Reset()         { *m = Balance{} }
func (m *Balance) String() string { return proto.CompactTextString(m) }
func (*Balance) ProtoMessage()    {}
func (*Balance) Descriptor() ([]byte, []int) {
	return fileDescriptor_d281657d6c91cb92, []int{1}
}
func (m *Balance) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Balance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Balance.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Balance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Balance.Merge(m, src)
}
func (m *Balance) XXX_Size() int {
	return m.Size()
}
func (m *Balance) XXX_DiscardUnknown() {
	xxx_messageInfo_Balance.DiscardUnknown(m)
}

var xxx_messageInfo_Balance proto.InternalMessageInfo

func (m *Balance) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Balance) GetCoins() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Coins
	}
	return nil
}

// PendingTokenUpgrade stores the version of pending token upgrade.
type PendingTokenUpgrade struct {
	Denom   string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Version uint32 `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
}

func (m *PendingTokenUpgrade) Reset()         { *m = PendingTokenUpgrade{} }
func (m *PendingTokenUpgrade) String() string { return proto.CompactTextString(m) }
func (*PendingTokenUpgrade) ProtoMessage()    {}
func (*PendingTokenUpgrade) Descriptor() ([]byte, []int) {
	return fileDescriptor_d281657d6c91cb92, []int{2}
}
func (m *PendingTokenUpgrade) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PendingTokenUpgrade) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PendingTokenUpgrade.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PendingTokenUpgrade) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PendingTokenUpgrade.Merge(m, src)
}
func (m *PendingTokenUpgrade) XXX_Size() int {
	return m.Size()
}
func (m *PendingTokenUpgrade) XXX_DiscardUnknown() {
	xxx_messageInfo_PendingTokenUpgrade.DiscardUnknown(m)
}

var xxx_messageInfo_PendingTokenUpgrade proto.InternalMessageInfo

func (m *PendingTokenUpgrade) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *PendingTokenUpgrade) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "coreum.asset.ft.v1.GenesisState")
	proto.RegisterType((*Balance)(nil), "coreum.asset.ft.v1.Balance")
	proto.RegisterType((*PendingTokenUpgrade)(nil), "coreum.asset.ft.v1.PendingTokenUpgrade")
}

func init() { proto.RegisterFile("coreum/asset/ft/v1/genesis.proto", fileDescriptor_d281657d6c91cb92) }

var fileDescriptor_d281657d6c91cb92 = []byte{
	// 466 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0x8d, 0x9b, 0x26, 0x15, 0x5b, 0x3e, 0xa4, 0x6d, 0x84, 0x4c, 0x90, 0x9c, 0x28, 0x17, 0x72,
	0x61, 0x97, 0xb4, 0x12, 0x70, 0x4e, 0x05, 0x48, 0x9c, 0xaa, 0x50, 0x2e, 0x5c, 0xa2, 0xb5, 0x3d,
	0x71, 0x57, 0xad, 0x77, 0x2d, 0xcf, 0xc6, 0x7c, 0xfc, 0x00, 0xc4, 0x91, 0xdf, 0xc1, 0x2f, 0xe9,
	0xb1, 0x47, 0x4e, 0x80, 0x92, 0x3f, 0x82, 0xbc, 0xbb, 0x21, 0x91, 0xea, 0x03, 0x27, 0xef, 0xee,
	0xbc, 0xf7, 0xe6, 0x79, 0xde, 0x90, 0x61, 0xa2, 0x4b, 0x58, 0xe6, 0x5c, 0x20, 0x82, 0xe1, 0x0b,
	0xc3, 0xab, 0x09, 0xcf, 0x40, 0x01, 0x4a, 0x64, 0x45, 0xa9, 0x8d, 0xa6, 0xd4, 0x21, 0x98, 0x45,
	0xb0, 0x85, 0x61, 0xd5, 0xa4, 0x3f, 0x68, 0x60, 0x15, 0xa2, 0x14, 0xb9, 0x27, 0xf5, 0xa3, 0x06,
	0x80, 0xd1, 0x97, 0xa0, 0xb6, 0x75, 0xcc, 0x35, 0xf2, 0x58, 0x20, 0xf0, 0x6a, 0x12, 0x83, 0x11,
	0x13, 0x9e, 0x68, 0xb9, 0xa9, 0xf7, 0x32, 0x9d, 0x69, 0x7b, 0xe4, 0xf5, 0xc9, 0xbd, 0x8e, 0xbe,
	0xb5, 0xc9, 0xdd, 0x37, 0xce, 0xdc, 0x3b, 0x23, 0x0c, 0xd0, 0x97, 0xa4, 0xeb, 0xda, 0x86, 0xc1,
	0x30, 0x18, 0x1f, 0x1e, 0xf7, 0xd9, 0x6d, 0xb3, 0xec, 0xcc, 0x22, 0xa6, 0xfb, 0xd7, 0xbf, 0x06,
	0xad, 0x99, 0xc7, 0xd3, 0x17, 0xa4, 0x6b, 0xfd, 0x60, 0xb8, 0x37, 0x6c, 0x8f, 0x0f, 0x8f, 0x1f,
	0x35, 0x31, 0xcf, 0x6b, 0xc4, 0x86, 0xe8, 0xe0, 0xf4, 0x2d, 0x79, 0xb0, 0x28, 0xf5, 0x17, 0x50,
	0xf3, 0x58, 0x5c, 0x09, 0x95, 0x00, 0x86, 0x6d, 0xab, 0xf0, 0xb8, 0x49, 0x61, 0xea, 0x30, 0x5e,
	0xe3, 0xbe, 0x63, 0xfa, 0x47, 0xa4, 0xe7, 0xa4, 0xf7, 0xf1, 0x42, 0x1a, 0xb8, 0x92, 0x68, 0x20,
	0xdd, 0x0a, 0xee, 0xff, 0xaf, 0xe0, 0xd1, 0x0e, 0xfd, 0x9f, 0x6a, 0x42, 0x1e, 0x16, 0xa0, 0x52,
	0xa9, 0xb2, 0xb9, 0xf5, 0x3c, 0x5f, 0x16, 0x59, 0x29, 0x52, 0xc0, 0xb0, 0x63, 0x75, 0x9f, 0x34,
	0x0e, 0xc9, 0x31, 0xec, 0x1f, 0xbf, 0x77, 0x78, 0xdf, 0xa3, 0x57, 0xdc, 0x2e, 0xe1, 0xe8, 0x6b,
	0x40, 0x0e, 0x7c, 0x47, 0x1a, 0x92, 0x03, 0x91, 0xa6, 0x25, 0xa0, 0x8b, 0xe1, 0xce, 0x6c, 0x73,
	0xa5, 0x82, 0x74, 0xea, 0x50, 0x77, 0x87, 0x5c, 0xc7, 0xce, 0xea, 0xd8, 0x99, 0x8f, 0x9d, 0x9d,
	0x6a, 0xa9, 0xa6, 0xcf, 0xea, 0x5e, 0x3f, 0x7e, 0x0f, 0xc6, 0x99, 0x34, 0x17, 0xcb, 0x98, 0x25,
	0x3a, 0xe7, 0x7e, 0x47, 0xdc, 0xe7, 0x29, 0xa6, 0x97, 0xdc, 0x7c, 0x2e, 0x00, 0x2d, 0x01, 0x67,
	0x4e, 0x79, 0xf4, 0x8a, 0x1c, 0x35, 0x78, 0xa7, 0x3d, 0xd2, 0x49, 0x41, 0xe9, 0xdc, 0x3b, 0x72,
	0x97, 0xda, 0x69, 0x05, 0x25, 0x4a, 0xad, 0xc2, 0xbd, 0x61, 0x30, 0xbe, 0x37, 0xdb, 0x5c, 0xa7,
	0x67, 0xd7, 0xab, 0x28, 0xb8, 0x59, 0x45, 0xc1, 0x9f, 0x55, 0x14, 0x7c, 0x5f, 0x47, 0xad, 0x9b,
	0x75, 0xd4, 0xfa, 0xb9, 0x8e, 0x5a, 0x1f, 0x9e, 0xef, 0x38, 0x3a, 0xb5, 0x83, 0x7b, 0xad, 0x97,
	0x2a, 0x15, 0x46, 0x6a, 0xc5, 0xfd, 0x9a, 0x57, 0x27, 0xfc, 0xd3, 0x76, 0xd7, 0xad, 0xcb, 0xb8,
	0x6b, 0x77, 0xf6, 0xe4, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb2, 0xea, 0x7e, 0x1c, 0x62, 0x03,
	0x00, 0x00,
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
	if len(m.PendingTokenUpgrades) > 0 {
		for iNdEx := len(m.PendingTokenUpgrades) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.PendingTokenUpgrades[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.WhitelistedBalances) > 0 {
		for iNdEx := len(m.WhitelistedBalances) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.WhitelistedBalances[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.FrozenBalances) > 0 {
		for iNdEx := len(m.FrozenBalances) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.FrozenBalances[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Tokens) > 0 {
		for iNdEx := len(m.Tokens) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Tokens[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *Balance) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Balance) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Balance) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Coins) > 0 {
		for iNdEx := len(m.Coins) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Coins[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PendingTokenUpgrade) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PendingTokenUpgrade) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PendingTokenUpgrade) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Version != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Version))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
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
	if len(m.Tokens) > 0 {
		for _, e := range m.Tokens {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.FrozenBalances) > 0 {
		for _, e := range m.FrozenBalances {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.WhitelistedBalances) > 0 {
		for _, e := range m.WhitelistedBalances {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.PendingTokenUpgrades) > 0 {
		for _, e := range m.PendingTokenUpgrades {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *Balance) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Coins) > 0 {
		for _, e := range m.Coins {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *PendingTokenUpgrade) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Version != 0 {
		n += 1 + sovGenesis(uint64(m.Version))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Tokens", wireType)
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
			m.Tokens = append(m.Tokens, Token{})
			if err := m.Tokens[len(m.Tokens)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FrozenBalances", wireType)
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
			m.FrozenBalances = append(m.FrozenBalances, Balance{})
			if err := m.FrozenBalances[len(m.FrozenBalances)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WhitelistedBalances", wireType)
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
			m.WhitelistedBalances = append(m.WhitelistedBalances, Balance{})
			if err := m.WhitelistedBalances[len(m.WhitelistedBalances)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PendingTokenUpgrades", wireType)
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
			m.PendingTokenUpgrades = append(m.PendingTokenUpgrades, PendingTokenUpgrade{})
			if err := m.PendingTokenUpgrades[len(m.PendingTokenUpgrades)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *Balance) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: Balance: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Balance: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Coins", wireType)
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
			m.Coins = append(m.Coins, types.Coin{})
			if err := m.Coins[len(m.Coins)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *PendingTokenUpgrade) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: PendingTokenUpgrade: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PendingTokenUpgrade: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
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
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= uint32(b&0x7F) << shift
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
