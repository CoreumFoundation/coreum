// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/customparams/v1/params.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/gogo/protobuf/gogoproto"
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

// StakingParams defines the set of additional staking params for the staking module wrapper.
type StakingParams struct {
	// min_self_delegation is the validators global self declared minimum for delegation.
	MinSelfDelegation github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,1,opt,name=min_self_delegation,json=minSelfDelegation,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"min_self_delegation" yaml:"min_self_delegation"`
}

func (m *StakingParams) Reset()         { *m = StakingParams{} }
func (m *StakingParams) String() string { return proto.CompactTextString(m) }
func (*StakingParams) ProtoMessage()    {}
func (*StakingParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_957be068a77b113f, []int{0}
}
func (m *StakingParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StakingParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StakingParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StakingParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StakingParams.Merge(m, src)
}
func (m *StakingParams) XXX_Size() int {
	return m.Size()
}
func (m *StakingParams) XXX_DiscardUnknown() {
	xxx_messageInfo_StakingParams.DiscardUnknown(m)
}

var xxx_messageInfo_StakingParams proto.InternalMessageInfo

func init() {
	proto.RegisterType((*StakingParams)(nil), "coreum.customparams.v1.StakingParams")
}

func init() {
	proto.RegisterFile("coreum/customparams/v1/params.proto", fileDescriptor_957be068a77b113f)
}

var fileDescriptor_957be068a77b113f = []byte{
	// 252 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4e, 0xce, 0x2f, 0x4a,
	0x2d, 0xcd, 0xd5, 0x4f, 0x2e, 0x2d, 0x2e, 0xc9, 0xcf, 0x2d, 0x48, 0x2c, 0x4a, 0xcc, 0x2d, 0xd6,
	0x2f, 0x33, 0xd4, 0x87, 0xb0, 0xf4, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0xc4, 0x20, 0x8a, 0xf4,
	0x90, 0x15, 0xe9, 0x95, 0x19, 0x4a, 0x89, 0xa4, 0xe7, 0xa7, 0xe7, 0x83, 0x95, 0xe8, 0x83, 0x58,
	0x10, 0xd5, 0x4a, 0xbd, 0x8c, 0x5c, 0xbc, 0xc1, 0x25, 0x89, 0xd9, 0x99, 0x79, 0xe9, 0x01, 0x60,
	0xa5, 0x42, 0x35, 0x5c, 0xc2, 0xb9, 0x99, 0x79, 0xf1, 0xc5, 0xa9, 0x39, 0x69, 0xf1, 0x29, 0xa9,
	0x39, 0xa9, 0xe9, 0x89, 0x25, 0x99, 0xf9, 0x79, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x4e, 0x3e,
	0x27, 0xee, 0xc9, 0x33, 0xdc, 0xba, 0x27, 0xaf, 0x96, 0x9e, 0x59, 0x92, 0x51, 0x9a, 0xa4, 0x97,
	0x9c, 0x9f, 0xab, 0x9f, 0x9c, 0x5f, 0x9c, 0x9b, 0x5f, 0x0c, 0xa5, 0x74, 0x8b, 0x53, 0xb2, 0xf5,
	0x4b, 0x2a, 0x0b, 0x52, 0x8b, 0xf5, 0x3c, 0xf3, 0x4a, 0x3e, 0xdd, 0x93, 0x97, 0xaa, 0x4c, 0xcc,
	0xcd, 0xb1, 0x52, 0xc2, 0x62, 0xa4, 0x52, 0x90, 0x60, 0x6e, 0x66, 0x5e, 0x70, 0x6a, 0x4e, 0x9a,
	0x0b, 0x5c, 0xcc, 0x29, 0xe4, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92,
	0x63, 0x9c, 0xf0, 0x58, 0x8e, 0xe1, 0xc2, 0x63, 0x39, 0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2, 0xac,
	0x90, 0xac, 0x74, 0x06, 0x7b, 0xd1, 0x2d, 0xbf, 0x34, 0x2f, 0x05, 0xac, 0x4d, 0x1f, 0x1a, 0x30,
	0x65, 0x46, 0xfa, 0x15, 0xa8, 0xa1, 0x03, 0x76, 0x4a, 0x12, 0x1b, 0xd8, 0xb3, 0xc6, 0x80, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x56, 0x0e, 0x36, 0x4c, 0x41, 0x01, 0x00, 0x00,
}

func (m *StakingParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StakingParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StakingParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.MinSelfDelegation.Size()
		i -= size
		if _, err := m.MinSelfDelegation.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *StakingParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.MinSelfDelegation.Size()
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *StakingParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
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
			return fmt.Errorf("proto: StakingParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StakingParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinSelfDelegation", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinSelfDelegation.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
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
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
