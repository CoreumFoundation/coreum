// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/dex/v1/event.proto

package types

import (
	fmt "fmt"
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

// EventOrderCreated is emitted when the limit order is saved to the order book.
type EventOrderCreated struct {
	Order Order `protobuf:"bytes,1,opt,name=order,proto3" json:"order"`
}

func (m *EventOrderCreated) Reset()         { *m = EventOrderCreated{} }
func (m *EventOrderCreated) String() string { return proto.CompactTextString(m) }
func (*EventOrderCreated) ProtoMessage()    {}
func (*EventOrderCreated) Descriptor() ([]byte, []int) {
	return fileDescriptor_cecfe712f14d2a81, []int{0}
}
func (m *EventOrderCreated) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventOrderCreated) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventOrderCreated.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventOrderCreated) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventOrderCreated.Merge(m, src)
}
func (m *EventOrderCreated) XXX_Size() int {
	return m.Size()
}
func (m *EventOrderCreated) XXX_DiscardUnknown() {
	xxx_messageInfo_EventOrderCreated.DiscardUnknown(m)
}

var xxx_messageInfo_EventOrderCreated proto.InternalMessageInfo

func (m *EventOrderCreated) GetOrder() Order {
	if m != nil {
		return m.Order
	}
	return Order{}
}

// EventOrderReduced is emitted when the order is reduced during the matching.
type EventOrderReduced struct {
	// creator is order creator address.
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	// id is unique order ID.
	ID string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	// sent_coin is coin sent during matching.
	SentCoin github_com_cosmos_cosmos_sdk_types.Coin `protobuf:"bytes,3,opt,name=sent_coin,json=sentCoin,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Coin" json:"sent_coin"`
	// received_coin is coin received during matching.
	ReceivedCoin github_com_cosmos_cosmos_sdk_types.Coin `protobuf:"bytes,4,opt,name=received_coin,json=receivedCoin,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Coin" json:"received_coin"`
}

func (m *EventOrderReduced) Reset()         { *m = EventOrderReduced{} }
func (m *EventOrderReduced) String() string { return proto.CompactTextString(m) }
func (*EventOrderReduced) ProtoMessage()    {}
func (*EventOrderReduced) Descriptor() ([]byte, []int) {
	return fileDescriptor_cecfe712f14d2a81, []int{1}
}
func (m *EventOrderReduced) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventOrderReduced) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventOrderReduced.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventOrderReduced) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventOrderReduced.Merge(m, src)
}
func (m *EventOrderReduced) XXX_Size() int {
	return m.Size()
}
func (m *EventOrderReduced) XXX_DiscardUnknown() {
	xxx_messageInfo_EventOrderReduced.DiscardUnknown(m)
}

var xxx_messageInfo_EventOrderReduced proto.InternalMessageInfo

func (m *EventOrderReduced) GetCreator() string {
	if m != nil {
		return m.Creator
	}
	return ""
}

func (m *EventOrderReduced) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

// EventOrderClosed is emitted when the order is closed during matching or manually.
type EventOrderClosed struct {
	Order Order `protobuf:"bytes,1,opt,name=order,proto3" json:"order"`
}

func (m *EventOrderClosed) Reset()         { *m = EventOrderClosed{} }
func (m *EventOrderClosed) String() string { return proto.CompactTextString(m) }
func (*EventOrderClosed) ProtoMessage()    {}
func (*EventOrderClosed) Descriptor() ([]byte, []int) {
	return fileDescriptor_cecfe712f14d2a81, []int{2}
}
func (m *EventOrderClosed) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventOrderClosed) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventOrderClosed.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventOrderClosed) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventOrderClosed.Merge(m, src)
}
func (m *EventOrderClosed) XXX_Size() int {
	return m.Size()
}
func (m *EventOrderClosed) XXX_DiscardUnknown() {
	xxx_messageInfo_EventOrderClosed.DiscardUnknown(m)
}

var xxx_messageInfo_EventOrderClosed proto.InternalMessageInfo

func (m *EventOrderClosed) GetOrder() Order {
	if m != nil {
		return m.Order
	}
	return Order{}
}

func init() {
	proto.RegisterType((*EventOrderCreated)(nil), "coreum.dex.v1.EventOrderCreated")
	proto.RegisterType((*EventOrderReduced)(nil), "coreum.dex.v1.EventOrderReduced")
	proto.RegisterType((*EventOrderClosed)(nil), "coreum.dex.v1.EventOrderClosed")
}

func init() { proto.RegisterFile("coreum/dex/v1/event.proto", fileDescriptor_cecfe712f14d2a81) }

var fileDescriptor_cecfe712f14d2a81 = []byte{
	// 333 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x91, 0xcd, 0x4a, 0xc3, 0x40,
	0x10, 0xc7, 0x93, 0x58, 0xab, 0x5d, 0x2d, 0x68, 0x28, 0x12, 0x7b, 0x48, 0xa5, 0x17, 0xbd, 0xb8,
	0x6b, 0x15, 0x5f, 0xa0, 0x1f, 0x82, 0x28, 0x08, 0xc1, 0x93, 0x17, 0x69, 0x77, 0x87, 0xba, 0x68,
	0x33, 0x25, 0xd9, 0x84, 0xfa, 0x16, 0x3e, 0x56, 0x8f, 0x3d, 0x8a, 0x87, 0x22, 0xe9, 0xc9, 0xb7,
	0x90, 0xdd, 0x6d, 0xa1, 0x7a, 0xec, 0x69, 0xe7, 0xe3, 0xbf, 0xbf, 0x19, 0xe6, 0x4f, 0x8e, 0x39,
	0x26, 0x90, 0x8d, 0x98, 0x80, 0x09, 0xcb, 0x5b, 0x0c, 0x72, 0x88, 0x15, 0x1d, 0x27, 0xa8, 0xd0,
	0xaf, 0xda, 0x16, 0x15, 0x30, 0xa1, 0x79, 0xab, 0xfe, 0x4f, 0x89, 0x89, 0x80, 0xc4, 0x2a, 0xeb,
	0xb5, 0x21, 0x0e, 0xd1, 0x84, 0x4c, 0x47, 0xb6, 0xda, 0xec, 0x91, 0xc3, 0x9e, 0xc6, 0x3d, 0x68,
	0x65, 0x27, 0x81, 0xbe, 0x02, 0xe1, 0x5f, 0x90, 0x6d, 0xf3, 0x33, 0x70, 0x4f, 0xdc, 0xb3, 0xbd,
	0xcb, 0x1a, 0xfd, 0x33, 0x84, 0x1a, 0x6d, 0xbb, 0x34, 0x9d, 0x37, 0x9c, 0xc8, 0x0a, 0x9b, 0x3f,
	0xee, 0x3a, 0x27, 0x02, 0x91, 0x71, 0x10, 0x7e, 0x40, 0x76, 0xb8, 0x46, 0xa2, 0x25, 0x55, 0xa2,
	0x55, 0xea, 0x1f, 0x11, 0x4f, 0x8a, 0xc0, 0xd3, 0xc5, 0x76, 0xb9, 0x98, 0x37, 0xbc, 0xdb, 0x6e,
	0xe4, 0x49, 0xe1, 0xdf, 0x93, 0x4a, 0x0a, 0xb1, 0x7a, 0xe6, 0x28, 0xe3, 0x60, 0xcb, 0xb4, 0x99,
	0x9e, 0xf3, 0x35, 0x6f, 0x9c, 0x0e, 0xa5, 0x7a, 0xc9, 0x06, 0x94, 0xe3, 0x88, 0x71, 0x4c, 0x47,
	0x98, 0x2e, 0x9f, 0xf3, 0x54, 0xbc, 0x32, 0xf5, 0x3e, 0x86, 0x94, 0x76, 0x50, 0xc6, 0xd1, 0xae,
	0x26, 0xe8, 0xc8, 0x7f, 0x24, 0xd5, 0x04, 0x38, 0xc8, 0x1c, 0x84, 0x25, 0x96, 0x36, 0x23, 0xee,
	0xaf, 0x28, 0x3a, 0x6b, 0x76, 0xc9, 0xc1, 0xda, 0xc9, 0xde, 0x30, 0xdd, 0xe4, 0x62, 0xed, 0xbb,
	0x69, 0x11, 0xba, 0xb3, 0x22, 0x74, 0xbf, 0x8b, 0xd0, 0xfd, 0x58, 0x84, 0xce, 0x6c, 0x11, 0x3a,
	0x9f, 0x8b, 0xd0, 0x79, 0x6a, 0xad, 0xad, 0xd5, 0x31, 0x98, 0x1b, 0xcc, 0x62, 0xd1, 0x57, 0x12,
	0x63, 0xb6, 0xf4, 0x37, 0xbf, 0x66, 0x13, 0x63, 0xb2, 0xd9, 0x72, 0x50, 0x36, 0x66, 0x5e, 0xfd,
	0x06, 0x00, 0x00, 0xff, 0xff, 0x1c, 0xc7, 0x6a, 0x7a, 0x29, 0x02, 0x00, 0x00,
}

func (m *EventOrderCreated) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventOrderCreated) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventOrderCreated) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Order.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *EventOrderReduced) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventOrderReduced) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventOrderReduced) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.ReceivedCoin.Size()
		i -= size
		if _, err := m.ReceivedCoin.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	{
		size := m.SentCoin.Size()
		i -= size
		if _, err := m.SentCoin.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.ID) > 0 {
		i -= len(m.ID)
		copy(dAtA[i:], m.ID)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.ID)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Creator) > 0 {
		i -= len(m.Creator)
		copy(dAtA[i:], m.Creator)
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Creator)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventOrderClosed) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventOrderClosed) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventOrderClosed) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Order.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvent(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
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
func (m *EventOrderCreated) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Order.Size()
	n += 1 + l + sovEvent(uint64(l))
	return n
}

func (m *EventOrderReduced) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Creator)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	l = m.SentCoin.Size()
	n += 1 + l + sovEvent(uint64(l))
	l = m.ReceivedCoin.Size()
	n += 1 + l + sovEvent(uint64(l))
	return n
}

func (m *EventOrderClosed) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Order.Size()
	n += 1 + l + sovEvent(uint64(l))
	return n
}

func sovEvent(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEvent(x uint64) (n int) {
	return sovEvent(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventOrderCreated) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: EventOrderCreated: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventOrderCreated: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Order", wireType)
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
			if err := m.Order.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *EventOrderReduced) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: EventOrderReduced: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventOrderReduced: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Creator", wireType)
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
			m.Creator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SentCoin", wireType)
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
			if err := m.SentCoin.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReceivedCoin", wireType)
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
			if err := m.ReceivedCoin.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *EventOrderClosed) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: EventOrderClosed: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventOrderClosed: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Order", wireType)
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
			if err := m.Order.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
