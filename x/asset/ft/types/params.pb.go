// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/ft/v1/params.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params store gov manageable parameters.
type Params struct {
	// issue_fee is the fee burnt each time new token is issued
	IssueFee types.Coin `protobuf:"bytes,1,opt,name=issue_fee,json=issueFee,proto3" json:"issue_fee" yaml:"issue_fee"`
	// ibc_decision_timeout defines the end of the decision period for enabling IBC
	TokenUpgradeDecisionTimeout time.Time `protobuf:"bytes,2,opt,name=token_upgrade_decision_timeout,json=tokenUpgradeDecisionTimeout,proto3,stdtime" json:"token_upgrade_decision_timeout" yaml:"token_upgrade_decision_timeout"`
	// ibc_grace_period defines the period after which IBC is effectively enabled
	TokenUpgradeGracePeriod time.Duration `protobuf:"bytes,3,opt,name=token_upgrade_grace_period,json=tokenUpgradeGracePeriod,proto3,stdduration" json:"token_upgrade_grace_period" yaml:"token_upgrade_grace_period"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_b08ee2013666b045, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetIssueFee() types.Coin {
	if m != nil {
		return m.IssueFee
	}
	return types.Coin{}
}

func (m *Params) GetTokenUpgradeDecisionTimeout() time.Time {
	if m != nil {
		return m.TokenUpgradeDecisionTimeout
	}
	return time.Time{}
}

func (m *Params) GetTokenUpgradeGracePeriod() time.Duration {
	if m != nil {
		return m.TokenUpgradeGracePeriod
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "coreum.asset.ft.v1.Params")
}

func init() { proto.RegisterFile("coreum/asset/ft/v1/params.proto", fileDescriptor_b08ee2013666b045) }

var fileDescriptor_b08ee2013666b045 = []byte{
	// 396 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0xb1, 0x8b, 0xd4, 0x40,
	0x14, 0xc6, 0x33, 0x27, 0x1c, 0x1a, 0x1b, 0x09, 0x82, 0x31, 0xc2, 0x44, 0x03, 0x82, 0x8d, 0x33,
	0xc4, 0xeb, 0x2c, 0x73, 0xc7, 0x59, 0x09, 0xcb, 0x71, 0x36, 0x36, 0x61, 0x92, 0xbc, 0xc4, 0xc1,
	0x4b, 0x5e, 0xc8, 0xcc, 0x2c, 0xee, 0x1f, 0x60, 0xbf, 0x58, 0xf9, 0x27, 0x6d, 0xb9, 0xa5, 0xd5,
	0x2a, 0xbb, 0xff, 0x81, 0x8d, 0xad, 0x64, 0x32, 0xab, 0xeb, 0x22, 0xd7, 0x4d, 0xf8, 0x7e, 0xef,
	0x7d, 0xdf, 0x17, 0x9e, 0x1f, 0x97, 0x38, 0x80, 0x69, 0xb9, 0x50, 0x0a, 0x34, 0xaf, 0x35, 0x9f,
	0xa7, 0xbc, 0x17, 0x83, 0x68, 0x15, 0xeb, 0x07, 0xd4, 0x18, 0x04, 0x13, 0xc0, 0x2c, 0xc0, 0x6a,
	0xcd, 0xe6, 0x69, 0xf4, 0xb0, 0xc1, 0x06, 0xad, 0xcc, 0xc7, 0xd7, 0x44, 0x46, 0x71, 0x83, 0xd8,
	0xdc, 0x00, 0xb7, 0x5f, 0x85, 0xa9, 0xb9, 0x96, 0x2d, 0x28, 0x2d, 0xda, 0xde, 0x01, 0xf4, 0x18,
	0xa8, 0xcc, 0x20, 0xb4, 0xc4, 0x6e, 0xaf, 0x97, 0xa8, 0x5a, 0x54, 0xbc, 0x10, 0x0a, 0xf8, 0x3c,
	0x2d, 0x40, 0x8b, 0x94, 0x97, 0x28, 0x9d, 0x9e, 0xfc, 0x3a, 0xf1, 0x4f, 0x67, 0x36, 0x5b, 0x30,
	0xf3, 0xef, 0x49, 0xa5, 0x0c, 0xe4, 0x35, 0x40, 0x48, 0x9e, 0x92, 0x17, 0xf7, 0x5f, 0x3d, 0x66,
	0xd3, 0x38, 0x1b, 0xc7, 0x99, 0x1b, 0x67, 0xe7, 0x28, 0xbb, 0x2c, 0x5c, 0x6d, 0x62, 0xef, 0xe7,
	0x26, 0x7e, 0xb0, 0x10, 0xed, 0xcd, 0xeb, 0xe4, 0xcf, 0x64, 0x72, 0x75, 0xd7, 0xbe, 0x2f, 0x01,
	0x82, 0x2f, 0xc4, 0xa7, 0x1a, 0x3f, 0x42, 0x97, 0x9b, 0xbe, 0x19, 0x44, 0x05, 0x79, 0x05, 0xa5,
	0x54, 0x12, 0xbb, 0x7c, 0xec, 0x81, 0x46, 0x87, 0x27, 0xd6, 0x27, 0x62, 0x53, 0x0d, 0xb6, 0xaf,
	0xc1, 0xae, 0xf7, 0x3d, 0xb3, 0xd4, 0x19, 0x3d, 0x9f, 0x8c, 0x6e, 0xdf, 0x97, 0x2c, 0xbf, 0xc7,
	0xe4, 0xea, 0x89, 0x85, 0xde, 0x4d, 0xcc, 0x85, 0x43, 0xae, 0x27, 0x22, 0xf8, 0x4c, 0xfc, 0xe8,
	0xdf, 0x25, 0xcd, 0x20, 0x4a, 0xc8, 0x7b, 0x18, 0x24, 0x56, 0xe1, 0x1d, 0x57, 0xfc, 0x38, 0xd0,
	0x85, 0xfb, 0xaf, 0xd9, 0x4b, 0x97, 0xe7, 0xd9, 0xff, 0xf2, 0x1c, 0xae, 0x4a, 0xbe, 0x8e, 0x59,
	0x1e, 0x1d, 0x66, 0x79, 0x33, 0xca, 0x33, 0xab, 0x66, 0x6f, 0x57, 0x5b, 0x4a, 0xd6, 0x5b, 0x4a,
	0x7e, 0x6c, 0x29, 0x59, 0xee, 0xa8, 0xb7, 0xde, 0x51, 0xef, 0xdb, 0x8e, 0x7a, 0xef, 0xcf, 0x1a,
	0xa9, 0x3f, 0x98, 0x82, 0x95, 0xd8, 0xf2, 0x73, 0x7b, 0x29, 0x97, 0x68, 0xba, 0xca, 0xda, 0x73,
	0x77, 0x5b, 0x9f, 0xfe, 0x5e, 0x97, 0x5e, 0xf4, 0xa0, 0x8a, 0x53, 0x9b, 0xf4, 0xec, 0x77, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xa4, 0x64, 0xd7, 0x84, 0x7d, 0x02, 0x00, 0x00,
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.TokenUpgradeGracePeriod, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.TokenUpgradeGracePeriod):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintParams(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x1a
	n2, err2 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.TokenUpgradeDecisionTimeout, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.TokenUpgradeDecisionTimeout):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintParams(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	{
		size, err := m.IssueFee.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
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
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.IssueFee.Size()
	n += 1 + l + sovParams(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.TokenUpgradeDecisionTimeout)
	n += 1 + l + sovParams(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.TokenUpgradeGracePeriod)
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IssueFee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.IssueFee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TokenUpgradeDecisionTimeout", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.TokenUpgradeDecisionTimeout, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TokenUpgradeGracePeriod", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.TokenUpgradeGracePeriod, dAtA[iNdEx:postIndex]); err != nil {
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
