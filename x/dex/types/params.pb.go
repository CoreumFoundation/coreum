// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/dex/v1/params.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
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

// Params keeps gov manageable parameters.
type Params struct {
	// default_unified_ref_amount is the default approximate amount you need to buy 1USD, used to for tokens without custom value
	DefaultUnifiedRefAmount cosmossdk_io_math.LegacyDec `protobuf:"bytes,1,opt,name=default_unified_ref_amount,json=defaultUnifiedRefAmount,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"default_unified_ref_amount"`
	// price_tick_exponent is the exponent used for the price tick calculation
	PriceTickExponent int32 `protobuf:"varint,2,opt,name=price_tick_exponent,json=priceTickExponent,proto3" json:"price_tick_exponent,omitempty"`
	// max_orders_per_denom is the maximum number of orders per denom the user can have
	MaxOrdersPerDenom uint64 `protobuf:"varint,3,opt,name=max_orders_per_denom,json=maxOrdersPerDenom,proto3" json:"max_orders_per_denom,omitempty"`
	// order_reserve is the reserve required to save the order in the order book
	OrderReserve github_com_cosmos_cosmos_sdk_types.Coin `protobuf:"bytes,4,opt,name=order_reserve,json=orderReserve,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Coin" json:"order_reserve"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f339dad46d471ea, []int{0}
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

func (m *Params) GetPriceTickExponent() int32 {
	if m != nil {
		return m.PriceTickExponent
	}
	return 0
}

func (m *Params) GetMaxOrdersPerDenom() uint64 {
	if m != nil {
		return m.MaxOrdersPerDenom
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "coreum.dex.v1.Params")
}

func init() { proto.RegisterFile("coreum/dex/v1/params.proto", fileDescriptor_4f339dad46d471ea) }

var fileDescriptor_4f339dad46d471ea = []byte{
	// 388 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xc1, 0x6e, 0xd4, 0x30,
	0x10, 0x86, 0xd7, 0xa5, 0x54, 0x22, 0xd0, 0x43, 0x43, 0x25, 0x96, 0x45, 0x4a, 0x57, 0x70, 0x60,
	0x2f, 0xd8, 0x0a, 0x88, 0x07, 0x60, 0x5b, 0xb8, 0x80, 0x44, 0x15, 0xc1, 0x85, 0x8b, 0xf1, 0xda,
	0x93, 0xd4, 0x0a, 0xf6, 0x44, 0xb6, 0x13, 0xa5, 0x6f, 0xc1, 0x03, 0xf1, 0x00, 0x3d, 0xf6, 0x88,
	0x38, 0x54, 0x68, 0xf7, 0x45, 0xd0, 0xda, 0x39, 0x70, 0xf2, 0xc8, 0xff, 0xf8, 0x1b, 0xcf, 0xff,
	0x67, 0x0b, 0x89, 0x0e, 0x7a, 0xc3, 0x14, 0x8c, 0x6c, 0x28, 0x59, 0x27, 0x9c, 0x30, 0x9e, 0x76,
	0x0e, 0x03, 0xe6, 0xc7, 0x49, 0xa3, 0x0a, 0x46, 0x3a, 0x94, 0x8b, 0x42, 0xa2, 0x37, 0xe8, 0xd9,
	0x46, 0x78, 0x60, 0x43, 0xb9, 0x81, 0x20, 0x4a, 0x26, 0x51, 0xdb, 0xd4, 0xbe, 0x38, 0x6d, 0xb0,
	0xc1, 0x58, 0xb2, 0x7d, 0x95, 0x6e, 0x9f, 0xff, 0x3a, 0xc8, 0x8e, 0x2e, 0x23, 0x35, 0xff, 0x9e,
	0x2d, 0x14, 0xd4, 0xa2, 0xff, 0x11, 0x78, 0x6f, 0x75, 0xad, 0x41, 0x71, 0x07, 0x35, 0x17, 0x06,
	0x7b, 0x1b, 0xe6, 0x64, 0x49, 0x56, 0x0f, 0xd6, 0x2f, 0x6e, 0xee, 0xce, 0x66, 0x7f, 0xee, 0xce,
	0x9e, 0xa5, 0x61, 0x5e, 0xb5, 0x54, 0x23, 0x33, 0x22, 0x5c, 0xd1, 0x4f, 0xd0, 0x08, 0x79, 0x7d,
	0x01, 0xb2, 0x7a, 0x32, 0x61, 0xbe, 0x26, 0x4a, 0x05, 0xf5, 0xbb, 0xc8, 0xc8, 0x69, 0xf6, 0xb8,
	0x73, 0x5a, 0x02, 0x0f, 0x5a, 0xb6, 0x1c, 0xc6, 0x0e, 0x2d, 0xd8, 0x30, 0x3f, 0x58, 0x92, 0xd5,
	0xfd, 0xea, 0x24, 0x4a, 0x5f, 0xb4, 0x6c, 0xdf, 0x4f, 0x42, 0xce, 0xb2, 0x53, 0x23, 0x46, 0x8e,
	0x4e, 0x81, 0xf3, 0xbc, 0x03, 0xc7, 0x15, 0x58, 0x34, 0xf3, 0x7b, 0x4b, 0xb2, 0x3a, 0xac, 0x4e,
	0x8c, 0x18, 0x3f, 0x47, 0xe9, 0x12, 0xdc, 0xc5, 0x5e, 0xc8, 0x31, 0x3b, 0x8e, 0xcd, 0xdc, 0x81,
	0x07, 0x37, 0xc0, 0xfc, 0x70, 0x49, 0x56, 0x0f, 0x5f, 0x3f, 0xa5, 0xe9, 0xbb, 0x74, 0xef, 0x0d,
	0x9d, 0xbc, 0xa1, 0xe7, 0xa8, 0xed, 0x9a, 0x4d, 0x0b, 0xbd, 0x6c, 0x74, 0xb8, 0xea, 0x37, 0x54,
	0xa2, 0x61, 0x93, 0x91, 0xe9, 0x78, 0xe5, 0x55, 0xcb, 0xc2, 0x75, 0x07, 0x3e, 0x3e, 0xa8, 0x1e,
	0xc5, 0x01, 0x55, 0xe2, 0xaf, 0x3f, 0xde, 0x6c, 0x0b, 0x72, 0xbb, 0x2d, 0xc8, 0xdf, 0x6d, 0x41,
	0x7e, 0xee, 0x8a, 0xd9, 0xed, 0xae, 0x98, 0xfd, 0xde, 0x15, 0xb3, 0x6f, 0xe5, 0x7f, 0xc0, 0xf3,
	0x18, 0xd4, 0x07, 0xec, 0xad, 0x12, 0x41, 0xa3, 0x65, 0x53, 0xaa, 0xc3, 0x5b, 0x36, 0xc6, 0x68,
	0x23, 0x7f, 0x73, 0x14, 0x23, 0x79, 0xf3, 0x2f, 0x00, 0x00, 0xff, 0xff, 0xf8, 0x02, 0x83, 0x7f,
	0xf5, 0x01, 0x00, 0x00,
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
	{
		size := m.OrderReserve.Size()
		i -= size
		if _, err := m.OrderReserve.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if m.MaxOrdersPerDenom != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxOrdersPerDenom))
		i--
		dAtA[i] = 0x18
	}
	if m.PriceTickExponent != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.PriceTickExponent))
		i--
		dAtA[i] = 0x10
	}
	{
		size := m.DefaultUnifiedRefAmount.Size()
		i -= size
		if _, err := m.DefaultUnifiedRefAmount.MarshalTo(dAtA[i:]); err != nil {
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
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.DefaultUnifiedRefAmount.Size()
	n += 1 + l + sovParams(uint64(l))
	if m.PriceTickExponent != 0 {
		n += 1 + sovParams(uint64(m.PriceTickExponent))
	}
	if m.MaxOrdersPerDenom != 0 {
		n += 1 + sovParams(uint64(m.MaxOrdersPerDenom))
	}
	l = m.OrderReserve.Size()
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
				return fmt.Errorf("proto: wrong wireType = %d for field DefaultUnifiedRefAmount", wireType)
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
			if err := m.DefaultUnifiedRefAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PriceTickExponent", wireType)
			}
			m.PriceTickExponent = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PriceTickExponent |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxOrdersPerDenom", wireType)
			}
			m.MaxOrdersPerDenom = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxOrdersPerDenom |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrderReserve", wireType)
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
			if err := m.OrderReserve.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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