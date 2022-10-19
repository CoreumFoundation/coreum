// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/v1/airdrop.proto

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

type AirdropFungibleToken struct {
	Id            github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,1,opt,name=id,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"id"`
	Sender        string                                 `protobuf:"bytes,2,opt,name=sender,proto3" json:"sender,omitempty"`
	SnapshotId    github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,3,opt,name=snapshot_id,json=snapshotId,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"snapshot_id"`
	Height        int64                                  `protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
	Description   string                                 `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	RequiredDenom string                                 `protobuf:"bytes,6,opt,name=required_denom,json=requiredDenom,proto3" json:"required_denom,omitempty"`
	Offer         []types.DecCoin                        `protobuf:"bytes,7,rep,name=offer,proto3" json:"offer"`
}

func (m *AirdropFungibleToken) Reset()         { *m = AirdropFungibleToken{} }
func (m *AirdropFungibleToken) String() string { return proto.CompactTextString(m) }
func (*AirdropFungibleToken) ProtoMessage()    {}
func (*AirdropFungibleToken) Descriptor() ([]byte, []int) {
	return fileDescriptor_b237b9e335f8d18d, []int{0}
}
func (m *AirdropFungibleToken) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AirdropFungibleToken) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AirdropFungibleToken.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AirdropFungibleToken) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AirdropFungibleToken.Merge(m, src)
}
func (m *AirdropFungibleToken) XXX_Size() int {
	return m.Size()
}
func (m *AirdropFungibleToken) XXX_DiscardUnknown() {
	xxx_messageInfo_AirdropFungibleToken.DiscardUnknown(m)
}

var xxx_messageInfo_AirdropFungibleToken proto.InternalMessageInfo

func init() {
	proto.RegisterType((*AirdropFungibleToken)(nil), "coreum.asset.v1.AirdropFungibleToken")
}

func init() { proto.RegisterFile("coreum/asset/v1/airdrop.proto", fileDescriptor_b237b9e335f8d18d) }

var fileDescriptor_b237b9e335f8d18d = []byte{
	// 389 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0xc1, 0x8a, 0xd4, 0x40,
	0x10, 0x4d, 0x32, 0xbb, 0xa3, 0xf6, 0xa0, 0x42, 0x58, 0x24, 0x2c, 0x9a, 0x04, 0x41, 0x99, 0x8b,
	0xdd, 0x46, 0x2f, 0xe2, 0x41, 0x70, 0x76, 0x59, 0x58, 0x2f, 0x42, 0xf0, 0xe4, 0x65, 0x49, 0xd2,
	0x35, 0x49, 0x33, 0xa6, 0x2b, 0x76, 0x77, 0x06, 0xfd, 0x03, 0x8f, 0x7e, 0xc2, 0x7c, 0xce, 0x1c,
	0xe7, 0x28, 0x1e, 0x06, 0x99, 0xf1, 0xe0, 0x67, 0x48, 0x3a, 0x19, 0x98, 0xf3, 0x9e, 0x92, 0xaa,
	0x57, 0xfd, 0xea, 0xd5, 0xab, 0x22, 0x4f, 0x0a, 0x54, 0xd0, 0xd6, 0x2c, 0xd3, 0x1a, 0x0c, 0x5b,
	0x26, 0x2c, 0x13, 0x8a, 0x2b, 0x6c, 0x68, 0xa3, 0xd0, 0xa0, 0xff, 0xb0, 0x87, 0xa9, 0x85, 0xe9,
	0x32, 0x39, 0x3f, 0x2b, 0xb1, 0x44, 0x8b, 0xb1, 0xee, 0xaf, 0x2f, 0x3b, 0x0f, 0x0b, 0xd4, 0x35,
	0x6a, 0x96, 0x67, 0x1a, 0xd8, 0x32, 0xc9, 0xc1, 0x64, 0x09, 0x2b, 0x50, 0xc8, 0x1e, 0x7f, 0xfa,
	0xd7, 0x23, 0x67, 0xef, 0x7b, 0xe2, 0xab, 0x56, 0x96, 0x22, 0xff, 0x02, 0x9f, 0x70, 0x01, 0xd2,
	0x7f, 0x47, 0x3c, 0xc1, 0x03, 0x37, 0x76, 0xa7, 0xf7, 0x66, 0x74, 0xbd, 0x8d, 0x9c, 0xdf, 0xdb,
	0xe8, 0x79, 0x29, 0x4c, 0xd5, 0xe6, 0xb4, 0xc0, 0x9a, 0x0d, 0xbc, 0xfd, 0xe7, 0x85, 0xe6, 0x0b,
	0x66, 0xbe, 0x37, 0xa0, 0xe9, 0xb5, 0x34, 0xa9, 0x27, 0xb8, 0xff, 0x88, 0x8c, 0x35, 0x48, 0x0e,
	0x2a, 0xf0, 0x3a, 0x8e, 0x74, 0x88, 0xfc, 0x8f, 0x64, 0xa2, 0x65, 0xd6, 0xe8, 0x0a, 0xcd, 0x8d,
	0xe0, 0xc1, 0xe8, 0x56, 0x0d, 0xc8, 0x81, 0xe2, 0xda, 0x36, 0xaa, 0x40, 0x94, 0x95, 0x09, 0x4e,
	0x62, 0x77, 0x3a, 0x4a, 0x87, 0xc8, 0x8f, 0xc9, 0x84, 0x83, 0x2e, 0x94, 0x68, 0x8c, 0x40, 0x19,
	0x9c, 0x5a, 0x15, 0xc7, 0x29, 0xff, 0x19, 0x79, 0xa0, 0xe0, 0x6b, 0x2b, 0x14, 0xf0, 0x1b, 0x0e,
	0x12, 0xeb, 0x60, 0x6c, 0x8b, 0xee, 0x1f, 0xb2, 0x97, 0x5d, 0xd2, 0x7f, 0x43, 0x4e, 0x71, 0x3e,
	0x07, 0x15, 0xdc, 0x89, 0x47, 0xd3, 0xc9, 0xab, 0xc7, 0xb4, 0x97, 0x44, 0x3b, 0x4b, 0xe9, 0x60,
	0x29, 0xbd, 0x84, 0xe2, 0x02, 0x85, 0x9c, 0x9d, 0x74, 0x93, 0xa4, 0xfd, 0x83, 0xb7, 0x77, 0x7f,
	0xac, 0x22, 0xe7, 0xdf, 0x2a, 0x72, 0x66, 0x1f, 0xd6, 0xbb, 0xd0, 0xdd, 0xec, 0x42, 0xf7, 0xcf,
	0x2e, 0x74, 0x7f, 0xee, 0x43, 0x67, 0xb3, 0x0f, 0x9d, 0x5f, 0xfb, 0xd0, 0xf9, 0xfc, 0xf2, 0x68,
	0xe4, 0x0b, 0xbb, 0xd2, 0x2b, 0x6c, 0x25, 0xcf, 0x3a, 0x85, 0x6c, 0x38, 0x81, 0x6f, 0xc3, 0x11,
	0x58, 0x03, 0xf2, 0xb1, 0xdd, 0xdc, 0xeb, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x5d, 0x7e, 0xe4,
	0x6c, 0x21, 0x02, 0x00, 0x00,
}

func (m *AirdropFungibleToken) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AirdropFungibleToken) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AirdropFungibleToken) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Offer) > 0 {
		for iNdEx := len(m.Offer) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Offer[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintAirdrop(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x3a
		}
	}
	if len(m.RequiredDenom) > 0 {
		i -= len(m.RequiredDenom)
		copy(dAtA[i:], m.RequiredDenom)
		i = encodeVarintAirdrop(dAtA, i, uint64(len(m.RequiredDenom)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintAirdrop(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x2a
	}
	if m.Height != 0 {
		i = encodeVarintAirdrop(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x20
	}
	{
		size := m.SnapshotId.Size()
		i -= size
		if _, err := m.SnapshotId.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintAirdrop(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintAirdrop(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0x12
	}
	{
		size := m.Id.Size()
		i -= size
		if _, err := m.Id.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintAirdrop(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintAirdrop(dAtA []byte, offset int, v uint64) int {
	offset -= sovAirdrop(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *AirdropFungibleToken) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Id.Size()
	n += 1 + l + sovAirdrop(uint64(l))
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovAirdrop(uint64(l))
	}
	l = m.SnapshotId.Size()
	n += 1 + l + sovAirdrop(uint64(l))
	if m.Height != 0 {
		n += 1 + sovAirdrop(uint64(m.Height))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovAirdrop(uint64(l))
	}
	l = len(m.RequiredDenom)
	if l > 0 {
		n += 1 + l + sovAirdrop(uint64(l))
	}
	if len(m.Offer) > 0 {
		for _, e := range m.Offer {
			l = e.Size()
			n += 1 + l + sovAirdrop(uint64(l))
		}
	}
	return n
}

func sovAirdrop(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAirdrop(x uint64) (n int) {
	return sovAirdrop(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AirdropFungibleToken) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAirdrop
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
			return fmt.Errorf("proto: AirdropFungibleToken: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AirdropFungibleToken: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
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
				return ErrInvalidLengthAirdrop
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAirdrop
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Id.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
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
				return ErrInvalidLengthAirdrop
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAirdrop
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
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
				return ErrInvalidLengthAirdrop
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAirdrop
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SnapshotId.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
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
				return ErrInvalidLengthAirdrop
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAirdrop
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequiredDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
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
				return ErrInvalidLengthAirdrop
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAirdrop
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RequiredDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Offer", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAirdrop
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
				return ErrInvalidLengthAirdrop
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAirdrop
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Offer = append(m.Offer, types.DecCoin{})
			if err := m.Offer[len(m.Offer)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAirdrop(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAirdrop
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
func skipAirdrop(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAirdrop
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
					return 0, ErrIntOverflowAirdrop
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
					return 0, ErrIntOverflowAirdrop
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
				return 0, ErrInvalidLengthAirdrop
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAirdrop
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAirdrop
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAirdrop        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAirdrop          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAirdrop = fmt.Errorf("proto: unexpected end of group")
)
