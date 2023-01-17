// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/nft/v1/genesis.proto

package types

import (
	fmt "fmt"
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

// GenesisState defines the nftasset module's genesis state.
type GenesisState struct {
	// params defines all the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// class_definitions keep the non-fungible token class definitions state
	ClassDefinitions []ClassDefinition `protobuf:"bytes,2,rep,name=class_definitions,json=classDefinitions,proto3" json:"class_definitions"`
	FrozenNfts       []FrozenNFT       `protobuf:"bytes,3,rep,name=frozen_nfts,json=frozenNfts,proto3" json:"frozen_nfts"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_3abcf08d60f6fbfd, []int{0}
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

func (m *GenesisState) GetClassDefinitions() []ClassDefinition {
	if m != nil {
		return m.ClassDefinitions
	}
	return nil
}

func (m *GenesisState) GetFrozenNfts() []FrozenNFT {
	if m != nil {
		return m.FrozenNfts
	}
	return nil
}

type FrozenNFT struct {
	ClassID string   `protobuf:"bytes,1,opt,name=classID,proto3" json:"classID,omitempty"`
	NftIDs  []string `protobuf:"bytes,2,rep,name=nftIDs,proto3" json:"nftIDs,omitempty"`
}

func (m *FrozenNFT) Reset()         { *m = FrozenNFT{} }
func (m *FrozenNFT) String() string { return proto.CompactTextString(m) }
func (*FrozenNFT) ProtoMessage()    {}
func (*FrozenNFT) Descriptor() ([]byte, []int) {
	return fileDescriptor_3abcf08d60f6fbfd, []int{1}
}
func (m *FrozenNFT) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FrozenNFT) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FrozenNFT.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FrozenNFT) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FrozenNFT.Merge(m, src)
}
func (m *FrozenNFT) XXX_Size() int {
	return m.Size()
}
func (m *FrozenNFT) XXX_DiscardUnknown() {
	xxx_messageInfo_FrozenNFT.DiscardUnknown(m)
}

var xxx_messageInfo_FrozenNFT proto.InternalMessageInfo

func (m *FrozenNFT) GetClassID() string {
	if m != nil {
		return m.ClassID
	}
	return ""
}

func (m *FrozenNFT) GetNftIDs() []string {
	if m != nil {
		return m.NftIDs
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "coreum.asset.nft.v1.GenesisState")
	proto.RegisterType((*FrozenNFT)(nil), "coreum.asset.nft.v1.FrozenNFT")
}

func init() { proto.RegisterFile("coreum/asset/nft/v1/genesis.proto", fileDescriptor_3abcf08d60f6fbfd) }

var fileDescriptor_3abcf08d60f6fbfd = []byte{
	// 335 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0xc1, 0x4e, 0xf2, 0x40,
	0x10, 0xc7, 0xdb, 0x8f, 0x2f, 0x18, 0x16, 0x0f, 0x5a, 0x8d, 0x69, 0x30, 0xae, 0x48, 0x3c, 0x70,
	0xda, 0x0d, 0xe8, 0xc5, 0x83, 0x17, 0x40, 0x0c, 0x17, 0x62, 0xd0, 0xc4, 0xc4, 0x0b, 0x59, 0xca,
	0x6e, 0x6d, 0x22, 0xbb, 0xa4, 0x3b, 0x10, 0xf5, 0x29, 0x7c, 0x2c, 0x8e, 0x1c, 0x3d, 0x19, 0x03,
	0x27, 0xdf, 0xc2, 0x74, 0x77, 0x25, 0x9a, 0xf4, 0xd6, 0xe9, 0xfc, 0xe6, 0x37, 0xff, 0x76, 0xd0,
	0x49, 0xa4, 0x52, 0x3e, 0x9b, 0x50, 0xa6, 0x35, 0x07, 0x2a, 0x05, 0xd0, 0x79, 0x83, 0xc6, 0x5c,
	0x72, 0x9d, 0x68, 0x32, 0x4d, 0x15, 0xa8, 0x60, 0xcf, 0x22, 0xc4, 0x20, 0x44, 0x0a, 0x20, 0xf3,
	0x46, 0x65, 0x3f, 0x56, 0xb1, 0x32, 0x7d, 0x9a, 0x3d, 0x59, 0xb4, 0x52, 0xcd, 0xb3, 0x4d, 0x59,
	0xca, 0x26, 0x4e, 0x56, 0x39, 0xca, 0x23, 0x32, 0xa7, 0x69, 0xd7, 0xbe, 0x7c, 0xb4, 0x7d, 0x6d,
	0xb7, 0xdf, 0x02, 0x03, 0x1e, 0x5c, 0xa0, 0xa2, 0x9d, 0x0f, 0xfd, 0xaa, 0x5f, 0x2f, 0x37, 0x0f,
	0x49, 0x4e, 0x1a, 0x72, 0x63, 0x90, 0xd6, 0xff, 0xc5, 0xc7, 0xb1, 0x37, 0x70, 0x03, 0xc1, 0x3d,
	0xda, 0x8d, 0x9e, 0x98, 0xd6, 0xc3, 0x31, 0x17, 0x89, 0x4c, 0x20, 0x51, 0x52, 0x87, 0xff, 0xaa,
	0x85, 0x7a, 0xb9, 0x79, 0x9a, 0x6b, 0x69, 0x67, 0x74, 0x67, 0x03, 0x3b, 0xdd, 0x4e, 0xf4, 0xf7,
	0xb5, 0x0e, 0xae, 0x50, 0x59, 0xa4, 0xea, 0x95, 0xcb, 0xa1, 0x14, 0xa0, 0xc3, 0x82, 0x51, 0xe2,
	0x5c, 0x65, 0xd7, 0x70, 0xfd, 0xee, 0x9d, 0x93, 0x21, 0x3b, 0xd8, 0x17, 0xa0, 0x6b, 0x97, 0xa8,
	0xb4, 0x69, 0x07, 0x21, 0xda, 0x32, 0x7b, 0x7a, 0x1d, 0xf3, 0xa1, 0xa5, 0xc1, 0x4f, 0x19, 0x1c,
	0xa0, 0xa2, 0x14, 0xd0, 0xeb, 0xd8, 0xec, 0xa5, 0x81, 0xab, 0x5a, 0xfd, 0xc5, 0x0a, 0xfb, 0xcb,
	0x15, 0xf6, 0x3f, 0x57, 0xd8, 0x7f, 0x5b, 0x63, 0x6f, 0xb9, 0xc6, 0xde, 0xfb, 0x1a, 0x7b, 0x0f,
	0xe7, 0x71, 0x02, 0x8f, 0xb3, 0x11, 0x89, 0xd4, 0x84, 0xb6, 0x4d, 0xa8, 0xae, 0x9a, 0xc9, 0x31,
	0xcb, 0xd2, 0x53, 0xf7, 0xff, 0x9f, 0x7f, 0x5d, 0x00, 0x5e, 0xa6, 0x5c, 0x8f, 0x8a, 0xe6, 0x02,
	0x67, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x19, 0x27, 0x50, 0xb7, 0x12, 0x02, 0x00, 0x00,
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
	if len(m.FrozenNfts) > 0 {
		for iNdEx := len(m.FrozenNfts) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.FrozenNfts[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ClassDefinitions) > 0 {
		for iNdEx := len(m.ClassDefinitions) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClassDefinitions[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *FrozenNFT) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FrozenNFT) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FrozenNFT) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.NftIDs) > 0 {
		for iNdEx := len(m.NftIDs) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.NftIDs[iNdEx])
			copy(dAtA[i:], m.NftIDs[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.NftIDs[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.ClassID) > 0 {
		i -= len(m.ClassID)
		copy(dAtA[i:], m.ClassID)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ClassID)))
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
	if len(m.ClassDefinitions) > 0 {
		for _, e := range m.ClassDefinitions {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.FrozenNfts) > 0 {
		for _, e := range m.FrozenNfts {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *FrozenNFT) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ClassID)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.NftIDs) > 0 {
		for _, s := range m.NftIDs {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
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
				return fmt.Errorf("proto: wrong wireType = %d for field ClassDefinitions", wireType)
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
			m.ClassDefinitions = append(m.ClassDefinitions, ClassDefinition{})
			if err := m.ClassDefinitions[len(m.ClassDefinitions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FrozenNfts", wireType)
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
			m.FrozenNfts = append(m.FrozenNfts, FrozenNFT{})
			if err := m.FrozenNfts[len(m.FrozenNfts)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *FrozenNFT) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: FrozenNFT: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FrozenNFT: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClassID", wireType)
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
			m.ClassID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NftIDs", wireType)
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
			m.NftIDs = append(m.NftIDs, string(dAtA[iNdEx:postIndex]))
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
