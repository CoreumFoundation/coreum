// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/snapshot/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type QueryPendingRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QueryPendingRequest) Reset()         { *m = QueryPendingRequest{} }
func (m *QueryPendingRequest) String() string { return proto.CompactTextString(m) }
func (*QueryPendingRequest) ProtoMessage()    {}
func (*QueryPendingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00a55b56744cc18c, []int{0}
}
func (m *QueryPendingRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPendingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPendingRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPendingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPendingRequest.Merge(m, src)
}
func (m *QueryPendingRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryPendingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPendingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPendingRequest proto.InternalMessageInfo

func (m *QueryPendingRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type QueryPendingResponse struct {
	Pending []SnapshotInfo `protobuf:"bytes,1,rep,name=pending,proto3" json:"pending"`
}

func (m *QueryPendingResponse) Reset()         { *m = QueryPendingResponse{} }
func (m *QueryPendingResponse) String() string { return proto.CompactTextString(m) }
func (*QueryPendingResponse) ProtoMessage()    {}
func (*QueryPendingResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_00a55b56744cc18c, []int{1}
}
func (m *QueryPendingResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPendingResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPendingResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPendingResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPendingResponse.Merge(m, src)
}
func (m *QueryPendingResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryPendingResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPendingResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPendingResponse proto.InternalMessageInfo

func (m *QueryPendingResponse) GetPending() []SnapshotInfo {
	if m != nil {
		return m.Pending
	}
	return nil
}

type QuerySnapshotsRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QuerySnapshotsRequest) Reset()         { *m = QuerySnapshotsRequest{} }
func (m *QuerySnapshotsRequest) String() string { return proto.CompactTextString(m) }
func (*QuerySnapshotsRequest) ProtoMessage()    {}
func (*QuerySnapshotsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00a55b56744cc18c, []int{2}
}
func (m *QuerySnapshotsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QuerySnapshotsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QuerySnapshotsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QuerySnapshotsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QuerySnapshotsRequest.Merge(m, src)
}
func (m *QuerySnapshotsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QuerySnapshotsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QuerySnapshotsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QuerySnapshotsRequest proto.InternalMessageInfo

func (m *QuerySnapshotsRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type QuerySnapshotsResponse struct {
	Snapshots []SnapshotInfo `protobuf:"bytes,1,rep,name=snapshots,proto3" json:"snapshots"`
}

func (m *QuerySnapshotsResponse) Reset()         { *m = QuerySnapshotsResponse{} }
func (m *QuerySnapshotsResponse) String() string { return proto.CompactTextString(m) }
func (*QuerySnapshotsResponse) ProtoMessage()    {}
func (*QuerySnapshotsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_00a55b56744cc18c, []int{3}
}
func (m *QuerySnapshotsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QuerySnapshotsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QuerySnapshotsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QuerySnapshotsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QuerySnapshotsResponse.Merge(m, src)
}
func (m *QuerySnapshotsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QuerySnapshotsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QuerySnapshotsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QuerySnapshotsResponse proto.InternalMessageInfo

func (m *QuerySnapshotsResponse) GetSnapshots() []SnapshotInfo {
	if m != nil {
		return m.Snapshots
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryPendingRequest)(nil), "coreum.snapshot.v1.QueryPendingRequest")
	proto.RegisterType((*QueryPendingResponse)(nil), "coreum.snapshot.v1.QueryPendingResponse")
	proto.RegisterType((*QuerySnapshotsRequest)(nil), "coreum.snapshot.v1.QuerySnapshotsRequest")
	proto.RegisterType((*QuerySnapshotsResponse)(nil), "coreum.snapshot.v1.QuerySnapshotsResponse")
}

func init() { proto.RegisterFile("coreum/snapshot/v1/query.proto", fileDescriptor_00a55b56744cc18c) }

var fileDescriptor_00a55b56744cc18c = []byte{
	// 380 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x4f, 0x4b, 0x02, 0x41,
	0x14, 0xdf, 0xb1, 0x3f, 0xe2, 0x74, 0x9b, 0x2c, 0x64, 0x89, 0x4d, 0x96, 0x42, 0x13, 0xda, 0x41,
	0xfd, 0x02, 0x61, 0x11, 0x74, 0x08, 0xca, 0x2e, 0xd1, 0x21, 0x58, 0xdd, 0x69, 0x5d, 0xd0, 0x99,
	0x75, 0x67, 0x56, 0x92, 0xe8, 0xd2, 0x35, 0x88, 0xa0, 0xef, 0xd0, 0x67, 0xf1, 0x28, 0x74, 0xe9,
	0x14, 0xa1, 0x7d, 0x90, 0x70, 0x76, 0x36, 0xc9, 0x36, 0xac, 0xdb, 0xdb, 0xfd, 0xbd, 0xf7, 0xfb,
	0xf3, 0xe6, 0x41, 0xa3, 0xc9, 0x02, 0x12, 0x76, 0x30, 0xa7, 0xb6, 0xcf, 0x5b, 0x4c, 0xe0, 0x5e,
	0x19, 0x77, 0x43, 0x12, 0xf4, 0x2d, 0x3f, 0x60, 0x82, 0x21, 0x14, 0xe1, 0x56, 0x8c, 0x5b, 0xbd,
	0xb2, 0x9e, 0x75, 0x99, 0xcb, 0x24, 0x8c, 0x27, 0x55, 0xd4, 0xa9, 0x6f, 0xb8, 0x8c, 0xb9, 0x6d,
	0x82, 0x6d, 0xdf, 0xc3, 0x36, 0xa5, 0x4c, 0xd8, 0xc2, 0x63, 0x94, 0x2b, 0x34, 0x49, 0x47, 0xf4,
	0x7d, 0xa2, 0x70, 0x13, 0xc3, 0xd5, 0xd3, 0x89, 0xec, 0x09, 0xa1, 0x8e, 0x47, 0xdd, 0x3a, 0xe9,
	0x86, 0x84, 0x0b, 0x94, 0x83, 0x69, 0xdb, 0x71, 0x02, 0xc2, 0x79, 0x0e, 0xe4, 0x41, 0x31, 0x53,
	0x8f, 0x3f, 0xcd, 0x73, 0x98, 0xfd, 0x3e, 0xc0, 0x7d, 0x46, 0x39, 0x41, 0x7b, 0x30, 0xed, 0x47,
	0xbf, 0x72, 0x20, 0xbf, 0x50, 0x5c, 0xa9, 0xe4, 0xad, 0x9f, 0x11, 0xac, 0x33, 0x55, 0x1f, 0xd1,
	0x2b, 0x56, 0x5b, 0x1c, 0xbc, 0x6d, 0x6a, 0xf5, 0x78, 0xcc, 0x2c, 0xc3, 0x35, 0xc9, 0x1c, 0xf7,
	0xf0, 0xf9, 0x66, 0x2e, 0xe1, 0xfa, 0xec, 0x88, 0xb2, 0x73, 0x00, 0x33, 0xb1, 0x2e, 0xff, 0xa7,
	0xa1, 0xe9, 0x60, 0xe5, 0x39, 0x05, 0x97, 0xa4, 0x00, 0xba, 0x07, 0x30, 0xad, 0x22, 0xa3, 0x42,
	0x12, 0x51, 0xc2, 0x16, 0xf5, 0xe2, 0xfc, 0xc6, 0xc8, 0xae, 0xb9, 0x7b, 0xf7, 0xf2, 0xf1, 0x94,
	0x2a, 0xa0, 0x6d, 0x9c, 0xf0, 0x5e, 0x6a, 0x41, 0xf8, 0x46, 0xc5, 0xbe, 0x45, 0x0f, 0x00, 0x66,
	0xbe, 0x32, 0xa3, 0x9d, 0x5f, 0x65, 0x66, 0x57, 0xa9, 0x97, 0xfe, 0xd2, 0xaa, 0x3c, 0x95, 0xa4,
	0xa7, 0x2d, 0x64, 0x26, 0x79, 0x6a, 0x7b, 0x5c, 0x4c, 0x0d, 0xd5, 0x8e, 0x07, 0x23, 0x03, 0x0c,
	0x47, 0x06, 0x78, 0x1f, 0x19, 0xe0, 0x71, 0x6c, 0x68, 0xc3, 0xb1, 0xa1, 0xbd, 0x8e, 0x0d, 0xed,
	0xa2, 0xea, 0x7a, 0xa2, 0x15, 0x36, 0xac, 0x26, 0xeb, 0xe0, 0x7d, 0xc9, 0x73, 0xc8, 0x42, 0xea,
	0xc8, 0x23, 0x8d, 0x89, 0xaf, 0xa7, 0xd4, 0xf2, 0x36, 0x1b, 0xcb, 0xf2, 0x38, 0xab, 0x9f, 0x01,
	0x00, 0x00, 0xff, 0xff, 0xfd, 0x9f, 0x40, 0x73, 0x26, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	Pending(ctx context.Context, in *QueryPendingRequest, opts ...grpc.CallOption) (*QueryPendingResponse, error)
	Snapshots(ctx context.Context, in *QuerySnapshotsRequest, opts ...grpc.CallOption) (*QuerySnapshotsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Pending(ctx context.Context, in *QueryPendingRequest, opts ...grpc.CallOption) (*QueryPendingResponse, error) {
	out := new(QueryPendingResponse)
	err := c.cc.Invoke(ctx, "/coreum.snapshot.v1.Query/Pending", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Snapshots(ctx context.Context, in *QuerySnapshotsRequest, opts ...grpc.CallOption) (*QuerySnapshotsResponse, error) {
	out := new(QuerySnapshotsResponse)
	err := c.cc.Invoke(ctx, "/coreum.snapshot.v1.Query/Snapshots", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	Pending(context.Context, *QueryPendingRequest) (*QueryPendingResponse, error)
	Snapshots(context.Context, *QuerySnapshotsRequest) (*QuerySnapshotsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Pending(ctx context.Context, req *QueryPendingRequest) (*QueryPendingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Pending not implemented")
}
func (*UnimplementedQueryServer) Snapshots(ctx context.Context, req *QuerySnapshotsRequest) (*QuerySnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Snapshots not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Pending_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPendingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Pending(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coreum.snapshot.v1.Query/Pending",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Pending(ctx, req.(*QueryPendingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Snapshots_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySnapshotsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Snapshots(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coreum.snapshot.v1.Query/Snapshots",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Snapshots(ctx, req.(*QuerySnapshotsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "coreum.snapshot.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Pending",
			Handler:    _Query_Pending_Handler,
		},
		{
			MethodName: "Snapshots",
			Handler:    _Query_Snapshots_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coreum/snapshot/v1/query.proto",
}

func (m *QueryPendingRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPendingRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPendingRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryPendingResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPendingResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPendingResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Pending) > 0 {
		for iNdEx := len(m.Pending) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Pending[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QuerySnapshotsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QuerySnapshotsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QuerySnapshotsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QuerySnapshotsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QuerySnapshotsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QuerySnapshotsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Snapshots) > 0 {
		for iNdEx := len(m.Snapshots) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Snapshots[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryPendingRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryPendingResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Pending) > 0 {
		for _, e := range m.Pending {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QuerySnapshotsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QuerySnapshotsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Snapshots) > 0 {
		for _, e := range m.Snapshots {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryPendingRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryPendingRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPendingRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryPendingResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryPendingResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPendingResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pending", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Pending = append(m.Pending, SnapshotInfo{})
			if err := m.Pending[len(m.Pending)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QuerySnapshotsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QuerySnapshotsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QuerySnapshotsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QuerySnapshotsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QuerySnapshotsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QuerySnapshotsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Snapshots", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Snapshots = append(m.Snapshots, SnapshotInfo{})
			if err := m.Snapshots[len(m.Snapshots)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
