// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/asset/ft/v1/upgradev1.proto

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

// MsgTokenUpgradeV1 is the message upgrading token to V1.
type MsgTokenUpgradeV1 struct {
	Sender     string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Denom      string `protobuf:"bytes,2,opt,name=denom,proto3" json:"denom,omitempty"`
	IbcEnabled bool   `protobuf:"varint,3,opt,name=ibc_enabled,json=ibcEnabled,proto3" json:"ibc_enabled,omitempty"`
}

func (m *MsgTokenUpgradeV1) Reset()         { *m = MsgTokenUpgradeV1{} }
func (m *MsgTokenUpgradeV1) String() string { return proto.CompactTextString(m) }
func (*MsgTokenUpgradeV1) ProtoMessage()    {}
func (*MsgTokenUpgradeV1) Descriptor() ([]byte, []int) {
	return fileDescriptor_78cded5dec668c74, []int{0}
}
func (m *MsgTokenUpgradeV1) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgTokenUpgradeV1) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgTokenUpgradeV1.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgTokenUpgradeV1) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgTokenUpgradeV1.Merge(m, src)
}
func (m *MsgTokenUpgradeV1) XXX_Size() int {
	return m.Size()
}
func (m *MsgTokenUpgradeV1) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgTokenUpgradeV1.DiscardUnknown(m)
}

var xxx_messageInfo_MsgTokenUpgradeV1 proto.InternalMessageInfo

// DelayedTokenUpgradeV1 is executed by the delay module when it's time to enable IBC.
type DelayedTokenUpgradeV1 struct {
	Denom      string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	IbcEnabled bool   `protobuf:"varint,2,opt,name=ibc_enabled,json=ibcEnabled,proto3" json:"ibc_enabled,omitempty"`
}

func (m *DelayedTokenUpgradeV1) Reset()         { *m = DelayedTokenUpgradeV1{} }
func (m *DelayedTokenUpgradeV1) String() string { return proto.CompactTextString(m) }
func (*DelayedTokenUpgradeV1) ProtoMessage()    {}
func (*DelayedTokenUpgradeV1) Descriptor() ([]byte, []int) {
	return fileDescriptor_78cded5dec668c74, []int{1}
}
func (m *DelayedTokenUpgradeV1) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DelayedTokenUpgradeV1) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DelayedTokenUpgradeV1.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DelayedTokenUpgradeV1) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DelayedTokenUpgradeV1.Merge(m, src)
}
func (m *DelayedTokenUpgradeV1) XXX_Size() int {
	return m.Size()
}
func (m *DelayedTokenUpgradeV1) XXX_DiscardUnknown() {
	xxx_messageInfo_DelayedTokenUpgradeV1.DiscardUnknown(m)
}

var xxx_messageInfo_DelayedTokenUpgradeV1 proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgTokenUpgradeV1)(nil), "coreum.asset.ft.v1.MsgTokenUpgradeV1")
	proto.RegisterType((*DelayedTokenUpgradeV1)(nil), "coreum.asset.ft.v1.DelayedTokenUpgradeV1")
}

func init() {
	proto.RegisterFile("coreum/asset/ft/v1/upgradev1.proto", fileDescriptor_78cded5dec668c74)
}

var fileDescriptor_78cded5dec668c74 = []byte{
	// 262 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4a, 0xce, 0x2f, 0x4a,
	0x2d, 0xcd, 0xd5, 0x4f, 0x2c, 0x2e, 0x4e, 0x2d, 0xd1, 0x4f, 0x2b, 0xd1, 0x2f, 0x33, 0xd4, 0x2f,
	0x2d, 0x48, 0x2f, 0x4a, 0x4c, 0x49, 0x2d, 0x33, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12,
	0x82, 0xa8, 0xd1, 0x03, 0xab, 0xd1, 0x4b, 0x2b, 0xd1, 0x2b, 0x33, 0x94, 0x12, 0x49, 0xcf, 0x4f,
	0xcf, 0x07, 0x4b, 0xeb, 0x83, 0x58, 0x10, 0x95, 0x4a, 0x49, 0x5c, 0x82, 0xbe, 0xc5, 0xe9, 0x21,
	0xf9, 0xd9, 0xa9, 0x79, 0xa1, 0x10, 0x43, 0xc2, 0x0c, 0x85, 0xc4, 0xb8, 0xd8, 0x8a, 0x53, 0xf3,
	0x52, 0x52, 0x8b, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xa0, 0x3c, 0x21, 0x11, 0x2e, 0xd6,
	0x94, 0xd4, 0xbc, 0xfc, 0x5c, 0x09, 0x26, 0xb0, 0x30, 0x84, 0x23, 0x24, 0xcf, 0xc5, 0x9d, 0x99,
	0x94, 0x1c, 0x9f, 0x9a, 0x97, 0x98, 0x94, 0x93, 0x9a, 0x22, 0xc1, 0xac, 0xc0, 0xa8, 0xc1, 0x11,
	0xc4, 0x95, 0x99, 0x94, 0xec, 0x0a, 0x11, 0x51, 0xf2, 0xe3, 0x12, 0x75, 0x49, 0xcd, 0x49, 0xac,
	0x4c, 0x4d, 0x41, 0xb3, 0x07, 0x6e, 0x1e, 0x23, 0x1e, 0xf3, 0x98, 0xd0, 0xcd, 0x73, 0x0a, 0x3c,
	0xf1, 0x50, 0x8e, 0xe1, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63,
	0x9c, 0xf0, 0x58, 0x8e, 0xe1, 0xc2, 0x63, 0x39, 0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2, 0x8c, 0xd3,
	0x33, 0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x9d, 0xc1, 0xc1, 0xe0, 0x96, 0x5f,
	0x9a, 0x97, 0x92, 0x58, 0x92, 0x99, 0x9f, 0xa7, 0x0f, 0x0d, 0xbb, 0x0a, 0x44, 0xe8, 0x95, 0x54,
	0x16, 0xa4, 0x16, 0x27, 0xb1, 0x81, 0x43, 0xc3, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x65, 0x1a,
	0x81, 0xba, 0x5d, 0x01, 0x00, 0x00,
}

func (m *MsgTokenUpgradeV1) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgTokenUpgradeV1) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgTokenUpgradeV1) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.IbcEnabled {
		i--
		if m.IbcEnabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x18
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintUpgradev1(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintUpgradev1(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DelayedTokenUpgradeV1) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DelayedTokenUpgradeV1) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DelayedTokenUpgradeV1) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.IbcEnabled {
		i--
		if m.IbcEnabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x10
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintUpgradev1(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintUpgradev1(dAtA []byte, offset int, v uint64) int {
	offset -= sovUpgradev1(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgTokenUpgradeV1) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovUpgradev1(uint64(l))
	}
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovUpgradev1(uint64(l))
	}
	if m.IbcEnabled {
		n += 2
	}
	return n
}

func (m *DelayedTokenUpgradeV1) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovUpgradev1(uint64(l))
	}
	if m.IbcEnabled {
		n += 2
	}
	return n
}

func sovUpgradev1(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozUpgradev1(x uint64) (n int) {
	return sovUpgradev1(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgTokenUpgradeV1) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUpgradev1
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
			return fmt.Errorf("proto: MsgTokenUpgradeV1: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgTokenUpgradeV1: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUpgradev1
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
				return ErrInvalidLengthUpgradev1
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthUpgradev1
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUpgradev1
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
				return ErrInvalidLengthUpgradev1
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthUpgradev1
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcEnabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUpgradev1
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
			m.IbcEnabled = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipUpgradev1(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUpgradev1
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
func (m *DelayedTokenUpgradeV1) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUpgradev1
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
			return fmt.Errorf("proto: DelayedTokenUpgradeV1: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DelayedTokenUpgradeV1: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUpgradev1
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
				return ErrInvalidLengthUpgradev1
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthUpgradev1
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcEnabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUpgradev1
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
			m.IbcEnabled = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipUpgradev1(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUpgradev1
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
func skipUpgradev1(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowUpgradev1
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
					return 0, ErrIntOverflowUpgradev1
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
					return 0, ErrIntOverflowUpgradev1
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
				return 0, ErrInvalidLengthUpgradev1
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupUpgradev1
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthUpgradev1
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthUpgradev1        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowUpgradev1          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupUpgradev1 = fmt.Errorf("proto: unexpected end of group")
)
