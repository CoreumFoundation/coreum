// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/dex/v1/tx.proto

package types

import (
	context "context"
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// MsgPlaceOrder defines message to place an order on orderbook.
type MsgPlaceOrder struct {
	// sender is order creator address.
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// id is unique order ID.
	ID string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	// base_denom is base order denom.
	BaseDenom string `protobuf:"bytes,3,opt,name=base_denom,json=baseDenom,proto3" json:"base_denom,omitempty"`
	// quote_denom is quote order denom
	QuoteDenom string `protobuf:"bytes,4,opt,name=quote_denom,json=quoteDenom,proto3" json:"quote_denom,omitempty"`
	// price is value of one unit of the base_denom expressed in terms of the quote_denom.
	Price Price `protobuf:"bytes,5,opt,name=price,proto3,customtype=Price" json:"price"`
	// quantity is amount of the base base_denom being traded.
	Quantity cosmossdk_io_math.Int `protobuf:"bytes,6,opt,name=quantity,proto3,customtype=cosmossdk.io/math.Int" json:"quantity"`
	// side is order side.
	Side Side `protobuf:"varint,7,opt,name=side,proto3,enum=coreum.dex.v1.Side" json:"side,omitempty"`
}

func (m *MsgPlaceOrder) Reset()         { *m = MsgPlaceOrder{} }
func (m *MsgPlaceOrder) String() string { return proto.CompactTextString(m) }
func (*MsgPlaceOrder) ProtoMessage()    {}
func (*MsgPlaceOrder) Descriptor() ([]byte, []int) {
	return fileDescriptor_6b3181ef84525da2, []int{0}
}
func (m *MsgPlaceOrder) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgPlaceOrder) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgPlaceOrder.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgPlaceOrder) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgPlaceOrder.Merge(m, src)
}
func (m *MsgPlaceOrder) XXX_Size() int {
	return m.Size()
}
func (m *MsgPlaceOrder) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgPlaceOrder.DiscardUnknown(m)
}

var xxx_messageInfo_MsgPlaceOrder proto.InternalMessageInfo

type EmptyResponse struct {
}

func (m *EmptyResponse) Reset()         { *m = EmptyResponse{} }
func (m *EmptyResponse) String() string { return proto.CompactTextString(m) }
func (*EmptyResponse) ProtoMessage()    {}
func (*EmptyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_6b3181ef84525da2, []int{1}
}
func (m *EmptyResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EmptyResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EmptyResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EmptyResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmptyResponse.Merge(m, src)
}
func (m *EmptyResponse) XXX_Size() int {
	return m.Size()
}
func (m *EmptyResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_EmptyResponse.DiscardUnknown(m)
}

var xxx_messageInfo_EmptyResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgPlaceOrder)(nil), "coreum.dex.v1.MsgPlaceOrder")
	proto.RegisterType((*EmptyResponse)(nil), "coreum.dex.v1.EmptyResponse")
}

func init() { proto.RegisterFile("coreum/dex/v1/tx.proto", fileDescriptor_6b3181ef84525da2) }

var fileDescriptor_6b3181ef84525da2 = []byte{
	// 304 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x50, 0x3f, 0x4b, 0xc3, 0x40,
	0x14, 0xcf, 0x29, 0x16, 0x3c, 0x29, 0xd2, 0x50, 0x6a, 0x2d, 0x72, 0xd4, 0x4e, 0xa5, 0xc3, 0x1d,
	0xad, 0x4e, 0x8e, 0x8a, 0xe2, 0x52, 0x22, 0x19, 0xdd, 0xd2, 0xe4, 0x38, 0x03, 0x5e, 0x5e, 0xc8,
	0x5d, 0x42, 0xba, 0x3a, 0x3a, 0xf9, 0x51, 0xfa, 0x31, 0x3a, 0x76, 0x74, 0xd4, 0x64, 0xe8, 0xd7,
	0x90, 0x5c, 0x32, 0x98, 0x2e, 0xc7, 0xbb, 0xf7, 0x7b, 0xbf, 0x3f, 0xfc, 0xf0, 0xc0, 0x87, 0x84,
	0xa7, 0x92, 0x05, 0x3c, 0x67, 0xd9, 0x9c, 0xe9, 0x9c, 0xc6, 0x09, 0x68, 0xb0, 0xbb, 0xf5, 0x9e,
	0x06, 0x3c, 0xa7, 0xd9, 0x7c, 0xd4, 0xf3, 0x64, 0x18, 0x01, 0x33, 0x6f, 0x7d, 0x31, 0xba, 0x6c,
	0x33, 0x21, 0x09, 0x78, 0xd2, 0x40, 0x17, 0x3e, 0x28, 0x09, 0x8a, 0x49, 0x25, 0x2a, 0x48, 0x2a,
	0xd1, 0x00, 0x7d, 0x01, 0x02, 0xcc, 0xc8, 0xaa, 0xa9, 0xde, 0x4e, 0x32, 0xdc, 0x5d, 0x2a, 0xf1,
	0xf2, 0xee, 0xf9, 0xdc, 0xa9, 0x54, 0xec, 0x01, 0xee, 0x28, 0x1e, 0x05, 0x3c, 0x19, 0xa2, 0x31,
	0x9a, 0x9e, 0xba, 0xcd, 0xcf, 0x9e, 0xe1, 0x13, 0x63, 0x33, 0x3c, 0x1a, 0xa3, 0xe9, 0xd9, 0xa2,
	0x4f, 0x5b, 0x21, 0xa9, 0x21, 0xbb, 0xf5, 0xc9, 0xdd, 0xf5, 0xc7, 0x7e, 0x33, 0x6b, 0x88, 0x9f,
	0xfb, 0xcd, 0xac, 0x57, 0xe5, 0x6c, 0xd9, 0x4c, 0xce, 0x71, 0xf7, 0x51, 0xc6, 0x7a, 0xed, 0x72,
	0x15, 0x43, 0xa4, 0xf8, 0xc2, 0xc1, 0xc7, 0x4b, 0x25, 0xec, 0x67, 0x8c, 0xff, 0x85, 0xb9, 0x3a,
	0x70, 0x69, 0x69, 0x8c, 0x0e, 0xd1, 0x96, 0xe0, 0xbd, 0xb3, 0xfd, 0x25, 0xd6, 0xb6, 0x20, 0x68,
	0x57, 0x10, 0xf4, 0x53, 0x10, 0xf4, 0x55, 0x12, 0x6b, 0x57, 0x12, 0xeb, 0xbb, 0x24, 0xd6, 0xeb,
	0x5c, 0x84, 0xfa, 0x2d, 0x5d, 0x51, 0x1f, 0x24, 0x7b, 0x30, 0x2a, 0x4f, 0x90, 0x46, 0x81, 0xa7,
	0x43, 0x88, 0x58, 0xd3, 0x6e, 0x76, 0xcb, 0x72, 0x53, 0xb1, 0x5e, 0xc7, 0x5c, 0xad, 0x3a, 0xa6,
	0xb1, 0x9b, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x46, 0xd3, 0xbb, 0x19, 0xb7, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// PlaceOrder is a method to place an order on orderbook.
	PlaceOrder(ctx context.Context, in *MsgPlaceOrder, opts ...grpc.CallOption) (*EmptyResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) PlaceOrder(ctx context.Context, in *MsgPlaceOrder, opts ...grpc.CallOption) (*EmptyResponse, error) {
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, "/coreum.dex.v1.Msg/PlaceOrder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// PlaceOrder is a method to place an order on orderbook.
	PlaceOrder(context.Context, *MsgPlaceOrder) (*EmptyResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) PlaceOrder(ctx context.Context, req *MsgPlaceOrder) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PlaceOrder not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_PlaceOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPlaceOrder)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PlaceOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coreum.dex.v1.Msg/PlaceOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PlaceOrder(ctx, req.(*MsgPlaceOrder))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "coreum.dex.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PlaceOrder",
			Handler:    _Msg_PlaceOrder_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coreum/dex/v1/tx.proto",
}

func (m *MsgPlaceOrder) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgPlaceOrder) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgPlaceOrder) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Side != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Side))
		i--
		dAtA[i] = 0x38
	}
	{
		size := m.Quantity.Size()
		i -= size
		if _, err := m.Quantity.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	{
		size := m.Price.Size()
		i -= size
		if _, err := m.Price.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	if len(m.QuoteDenom) > 0 {
		i -= len(m.QuoteDenom)
		copy(dAtA[i:], m.QuoteDenom)
		i = encodeVarintTx(dAtA, i, uint64(len(m.QuoteDenom)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.BaseDenom) > 0 {
		i -= len(m.BaseDenom)
		copy(dAtA[i:], m.BaseDenom)
		i = encodeVarintTx(dAtA, i, uint64(len(m.BaseDenom)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ID) > 0 {
		i -= len(m.ID)
		copy(dAtA[i:], m.ID)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ID)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EmptyResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EmptyResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EmptyResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgPlaceOrder) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.BaseDenom)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.QuoteDenom)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Price.Size()
	n += 1 + l + sovTx(uint64(l))
	l = m.Quantity.Size()
	n += 1 + l + sovTx(uint64(l))
	if m.Side != 0 {
		n += 1 + sovTx(uint64(m.Side))
	}
	return n
}

func (m *EmptyResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgPlaceOrder) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgPlaceOrder: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgPlaceOrder: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BaseDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field QuoteDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.QuoteDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Price", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Price.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Quantity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Quantity.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Side", wireType)
			}
			m.Side = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Side |= Side(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *EmptyResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: EmptyResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EmptyResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
