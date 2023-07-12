// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/feemodel/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
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

// QueryMinGasPriceRequest is the request type for the Query/MinGasPrice RPC method.
type QueryMinGasPriceRequest struct {
}

func (m *QueryMinGasPriceRequest) Reset()         { *m = QueryMinGasPriceRequest{} }
func (m *QueryMinGasPriceRequest) String() string { return proto.CompactTextString(m) }
func (*QueryMinGasPriceRequest) ProtoMessage()    {}
func (*QueryMinGasPriceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2036651e57006ae, []int{0}
}
func (m *QueryMinGasPriceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMinGasPriceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMinGasPriceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMinGasPriceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMinGasPriceRequest.Merge(m, src)
}
func (m *QueryMinGasPriceRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryMinGasPriceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMinGasPriceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMinGasPriceRequest proto.InternalMessageInfo

// QueryMinGasPriceResponse is the response type for the Query/MinGasPrice RPC method.
type QueryMinGasPriceResponse struct {
	// min_gas_price is the current minimum gas price required by the network.
	MinGasPrice types.DecCoin `protobuf:"bytes,1,opt,name=min_gas_price,json=minGasPrice,proto3" json:"min_gas_price"`
}

func (m *QueryMinGasPriceResponse) Reset()         { *m = QueryMinGasPriceResponse{} }
func (m *QueryMinGasPriceResponse) String() string { return proto.CompactTextString(m) }
func (*QueryMinGasPriceResponse) ProtoMessage()    {}
func (*QueryMinGasPriceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2036651e57006ae, []int{1}
}
func (m *QueryMinGasPriceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMinGasPriceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMinGasPriceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMinGasPriceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMinGasPriceResponse.Merge(m, src)
}
func (m *QueryMinGasPriceResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryMinGasPriceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMinGasPriceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMinGasPriceResponse proto.InternalMessageInfo

func (m *QueryMinGasPriceResponse) GetMinGasPrice() types.DecCoin {
	if m != nil {
		return m.MinGasPrice
	}
	return types.DecCoin{}
}

type QueryRecommendedGasPriceRequest struct {
	AfterBlocks uint32 `protobuf:"varint,1,opt,name=after_blocks,json=afterBlocks,proto3" json:"after_blocks,omitempty"`
}

func (m *QueryRecommendedGasPriceRequest) Reset()         { *m = QueryRecommendedGasPriceRequest{} }
func (m *QueryRecommendedGasPriceRequest) String() string { return proto.CompactTextString(m) }
func (*QueryRecommendedGasPriceRequest) ProtoMessage()    {}
func (*QueryRecommendedGasPriceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2036651e57006ae, []int{2}
}
func (m *QueryRecommendedGasPriceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryRecommendedGasPriceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryRecommendedGasPriceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryRecommendedGasPriceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryRecommendedGasPriceRequest.Merge(m, src)
}
func (m *QueryRecommendedGasPriceRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryRecommendedGasPriceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryRecommendedGasPriceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryRecommendedGasPriceRequest proto.InternalMessageInfo

func (m *QueryRecommendedGasPriceRequest) GetAfterBlocks() uint32 {
	if m != nil {
		return m.AfterBlocks
	}
	return 0
}

type QueryRecommendedGasPriceResponse struct {
	Low  types.DecCoin `protobuf:"bytes,1,opt,name=low,proto3" json:"low"`
	Med  types.DecCoin `protobuf:"bytes,2,opt,name=med,proto3" json:"med"`
	High types.DecCoin `protobuf:"bytes,3,opt,name=high,proto3" json:"high"`
}

func (m *QueryRecommendedGasPriceResponse) Reset()         { *m = QueryRecommendedGasPriceResponse{} }
func (m *QueryRecommendedGasPriceResponse) String() string { return proto.CompactTextString(m) }
func (*QueryRecommendedGasPriceResponse) ProtoMessage()    {}
func (*QueryRecommendedGasPriceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2036651e57006ae, []int{3}
}
func (m *QueryRecommendedGasPriceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryRecommendedGasPriceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryRecommendedGasPriceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryRecommendedGasPriceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryRecommendedGasPriceResponse.Merge(m, src)
}
func (m *QueryRecommendedGasPriceResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryRecommendedGasPriceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryRecommendedGasPriceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryRecommendedGasPriceResponse proto.InternalMessageInfo

func (m *QueryRecommendedGasPriceResponse) GetLow() types.DecCoin {
	if m != nil {
		return m.Low
	}
	return types.DecCoin{}
}

func (m *QueryRecommendedGasPriceResponse) GetMed() types.DecCoin {
	if m != nil {
		return m.Med
	}
	return types.DecCoin{}
}

func (m *QueryRecommendedGasPriceResponse) GetHigh() types.DecCoin {
	if m != nil {
		return m.High
	}
	return types.DecCoin{}
}

// QueryParamsRequest defines the request type for querying x/feemodel parameters.
type QueryParamsRequest struct {
}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()    {}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2036651e57006ae, []int{4}
}
func (m *QueryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsRequest.Merge(m, src)
}
func (m *QueryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsRequest proto.InternalMessageInfo

// QueryParamsResponse defines the response type for querying x/feemodel parameters.
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()    {}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2036651e57006ae, []int{5}
}
func (m *QueryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsResponse.Merge(m, src)
}
func (m *QueryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsResponse proto.InternalMessageInfo

func (m *QueryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func init() {
	proto.RegisterType((*QueryMinGasPriceRequest)(nil), "coreum.feemodel.v1.QueryMinGasPriceRequest")
	proto.RegisterType((*QueryMinGasPriceResponse)(nil), "coreum.feemodel.v1.QueryMinGasPriceResponse")
	proto.RegisterType((*QueryRecommendedGasPriceRequest)(nil), "coreum.feemodel.v1.QueryRecommendedGasPriceRequest")
	proto.RegisterType((*QueryRecommendedGasPriceResponse)(nil), "coreum.feemodel.v1.QueryRecommendedGasPriceResponse")
	proto.RegisterType((*QueryParamsRequest)(nil), "coreum.feemodel.v1.QueryParamsRequest")
	proto.RegisterType((*QueryParamsResponse)(nil), "coreum.feemodel.v1.QueryParamsResponse")
}

func init() { proto.RegisterFile("coreum/feemodel/v1/query.proto", fileDescriptor_d2036651e57006ae) }

var fileDescriptor_d2036651e57006ae = []byte{
	// 515 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x94, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x86, 0xe3, 0xb6, 0xe4, 0xb0, 0xa1, 0x97, 0x6d, 0x25, 0x82, 0x15, 0x39, 0xad, 0x91, 0x80,
	0xaa, 0xc8, 0xab, 0x34, 0x15, 0xe2, 0x9c, 0x56, 0xe5, 0x54, 0x51, 0x72, 0xe4, 0x12, 0xad, 0xed,
	0xa9, 0x63, 0x91, 0xdd, 0x71, 0xbd, 0x4e, 0xa0, 0x07, 0x2e, 0x3c, 0x01, 0x52, 0x1f, 0x85, 0x77,
	0x40, 0x3d, 0x56, 0x70, 0xe1, 0x84, 0x50, 0xc2, 0x83, 0x20, 0xaf, 0x37, 0x4a, 0x43, 0x6d, 0x91,
	0xde, 0xa2, 0xfd, 0xe7, 0xff, 0xe7, 0xdb, 0xcc, 0x78, 0x89, 0x13, 0x60, 0x0a, 0x63, 0xc1, 0xce,
	0x01, 0x04, 0x86, 0x30, 0x62, 0x93, 0x0e, 0xbb, 0x18, 0x43, 0x7a, 0xe9, 0x25, 0x29, 0x66, 0x48,
	0x69, 0xa1, 0x7b, 0x73, 0xdd, 0x9b, 0x74, 0xec, 0xed, 0x08, 0x23, 0xd4, 0x32, 0xcb, 0x7f, 0x15,
	0x95, 0x76, 0x2b, 0x42, 0x8c, 0x46, 0xc0, 0x78, 0x12, 0x33, 0x2e, 0x25, 0x66, 0x3c, 0x8b, 0x51,
	0x2a, 0xa3, 0x3a, 0x01, 0x2a, 0x81, 0x8a, 0xf9, 0x5c, 0x01, 0x9b, 0x74, 0x7c, 0xc8, 0x78, 0x87,
	0x05, 0x18, 0x4b, 0xa3, 0xb7, 0x4b, 0x38, 0x12, 0x9e, 0x72, 0x61, 0x02, 0xdc, 0xc7, 0xe4, 0xd1,
	0xdb, 0x9c, 0xeb, 0x34, 0x96, 0xaf, 0xb9, 0x3a, 0x4b, 0xe3, 0x00, 0xfa, 0x70, 0x31, 0x06, 0x95,
	0xb9, 0x3e, 0x69, 0xde, 0x95, 0x54, 0x82, 0x52, 0x01, 0x3d, 0x21, 0x9b, 0x22, 0x96, 0x83, 0x88,
	0xab, 0x41, 0x92, 0x0b, 0x4d, 0x6b, 0xc7, 0x7a, 0xde, 0x38, 0x68, 0x79, 0x05, 0x8f, 0x97, 0xf3,
	0x78, 0x86, 0xc7, 0x3b, 0x86, 0xe0, 0x08, 0x63, 0xd9, 0xdb, 0xb8, 0xfe, 0xd5, 0xae, 0xf5, 0x1b,
	0x62, 0x91, 0xe7, 0x1e, 0x93, 0xb6, 0xee, 0xd1, 0x87, 0x00, 0x85, 0x00, 0x19, 0x42, 0xf8, 0x0f,
	0x06, 0xdd, 0x25, 0x0f, 0xf9, 0x79, 0x06, 0xe9, 0xc0, 0x1f, 0x61, 0xf0, 0x5e, 0xe9, 0x4e, 0x9b,
	0xfd, 0x86, 0x3e, 0xeb, 0xe9, 0x23, 0xf7, 0x9b, 0x45, 0x76, 0xaa, 0x63, 0x0c, 0xf2, 0x21, 0x59,
	0x1f, 0xe1, 0x87, 0x7b, 0x80, 0xe6, 0xe5, 0xb9, 0x4b, 0x40, 0xd8, 0x5c, 0x5b, 0xdd, 0x25, 0x20,
	0xa4, 0x2f, 0xc9, 0xc6, 0x30, 0x8e, 0x86, 0xcd, 0xf5, 0x95, 0x6d, 0xba, 0xde, 0xdd, 0x26, 0x54,
	0xdf, 0xe3, 0x4c, 0x8f, 0x68, 0x3e, 0x88, 0x37, 0x64, 0x6b, 0xe9, 0xd4, 0x5c, 0xe8, 0x15, 0xa9,
	0x17, 0xa3, 0x34, 0x77, 0xb2, 0xbd, 0xbb, 0x4b, 0xe5, 0x15, 0x1e, 0xd3, 0xc4, 0xd4, 0x1f, 0x7c,
	0x5f, 0x27, 0x0f, 0x74, 0x22, 0xbd, 0xb2, 0x48, 0xe3, 0xd6, 0x7c, 0xe9, 0x7e, 0x59, 0x46, 0xc5,
	0x82, 0xd8, 0x2f, 0x56, 0x2b, 0x2e, 0x70, 0xdd, 0xbd, 0xcf, 0x3f, 0xfe, 0x5c, 0xad, 0x3d, 0xa1,
	0xbb, 0xac, 0x64, 0x27, 0x97, 0x96, 0x89, 0x7e, 0xb5, 0xc8, 0x56, 0xc9, 0x28, 0x69, 0xb7, 0xb2,
	0x61, 0xf5, 0xfe, 0xd8, 0x87, 0xf7, 0x33, 0x19, 0xda, 0x8e, 0xa6, 0xdd, 0xa7, 0x7b, 0x65, 0xb4,
	0xe9, 0xc2, 0x78, 0x8b, 0xfa, 0x13, 0xa9, 0x17, 0xff, 0x36, 0x7d, 0x5a, 0xd9, 0x72, 0x69, 0xb0,
	0xf6, 0xb3, 0xff, 0xd6, 0x19, 0x1a, 0x57, 0xd3, 0xb4, 0xa8, 0xcd, 0x2a, 0xbf, 0xe7, 0xde, 0xe9,
	0xf5, 0xd4, 0xb1, 0x6e, 0xa6, 0x8e, 0xf5, 0x7b, 0xea, 0x58, 0x5f, 0x66, 0x4e, 0xed, 0x66, 0xe6,
	0xd4, 0x7e, 0xce, 0x9c, 0xda, 0xbb, 0x6e, 0x14, 0x67, 0xc3, 0xb1, 0xef, 0x05, 0x28, 0xd8, 0x91,
	0xf6, 0x9f, 0xe0, 0x58, 0x86, 0xfa, 0x21, 0x99, 0x07, 0x7e, 0x5c, 0x44, 0x66, 0x97, 0x09, 0x28,
	0xbf, 0xae, 0xdf, 0x87, 0xee, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9c, 0x6a, 0xfd, 0xf3, 0xca,
	0x04, 0x00, 0x00,
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
	// MinGasPrice queries the current minimum gas price required by the network.
	MinGasPrice(ctx context.Context, in *QueryMinGasPriceRequest, opts ...grpc.CallOption) (*QueryMinGasPriceResponse, error)
	// RecommendedGasPrice queries the recommended gas price for the next n blocks.
	RecommendedGasPrice(ctx context.Context, in *QueryRecommendedGasPriceRequest, opts ...grpc.CallOption) (*QueryRecommendedGasPriceResponse, error)
	// Params queries the parameters of x/feemodel module.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) MinGasPrice(ctx context.Context, in *QueryMinGasPriceRequest, opts ...grpc.CallOption) (*QueryMinGasPriceResponse, error) {
	out := new(QueryMinGasPriceResponse)
	err := c.cc.Invoke(ctx, "/coreum.feemodel.v1.Query/MinGasPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) RecommendedGasPrice(ctx context.Context, in *QueryRecommendedGasPriceRequest, opts ...grpc.CallOption) (*QueryRecommendedGasPriceResponse, error) {
	out := new(QueryRecommendedGasPriceResponse)
	err := c.cc.Invoke(ctx, "/coreum.feemodel.v1.Query/RecommendedGasPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/coreum.feemodel.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// MinGasPrice queries the current minimum gas price required by the network.
	MinGasPrice(context.Context, *QueryMinGasPriceRequest) (*QueryMinGasPriceResponse, error)
	// RecommendedGasPrice queries the recommended gas price for the next n blocks.
	RecommendedGasPrice(context.Context, *QueryRecommendedGasPriceRequest) (*QueryRecommendedGasPriceResponse, error)
	// Params queries the parameters of x/feemodel module.
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) MinGasPrice(ctx context.Context, req *QueryMinGasPriceRequest) (*QueryMinGasPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MinGasPrice not implemented")
}
func (*UnimplementedQueryServer) RecommendedGasPrice(ctx context.Context, req *QueryRecommendedGasPriceRequest) (*QueryRecommendedGasPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecommendedGasPrice not implemented")
}
func (*UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_MinGasPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMinGasPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MinGasPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coreum.feemodel.v1.Query/MinGasPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MinGasPrice(ctx, req.(*QueryMinGasPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_RecommendedGasPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRecommendedGasPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).RecommendedGasPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coreum.feemodel.v1.Query/RecommendedGasPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).RecommendedGasPrice(ctx, req.(*QueryRecommendedGasPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coreum.feemodel.v1.Query/Params",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "coreum.feemodel.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MinGasPrice",
			Handler:    _Query_MinGasPrice_Handler,
		},
		{
			MethodName: "RecommendedGasPrice",
			Handler:    _Query_RecommendedGasPrice_Handler,
		},
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coreum/feemodel/v1/query.proto",
}

func (m *QueryMinGasPriceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMinGasPriceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMinGasPriceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryMinGasPriceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMinGasPriceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMinGasPriceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.MinGasPrice.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryRecommendedGasPriceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryRecommendedGasPriceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryRecommendedGasPriceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.AfterBlocks != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.AfterBlocks))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryRecommendedGasPriceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryRecommendedGasPriceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryRecommendedGasPriceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.High.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.Med.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.Low.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
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
func (m *QueryMinGasPriceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryMinGasPriceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.MinGasPrice.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryRecommendedGasPriceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.AfterBlocks != 0 {
		n += 1 + sovQuery(uint64(m.AfterBlocks))
	}
	return n
}

func (m *QueryRecommendedGasPriceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Low.Size()
	n += 1 + l + sovQuery(uint64(l))
	l = m.Med.Size()
	n += 1 + l + sovQuery(uint64(l))
	l = m.High.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryMinGasPriceRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMinGasPriceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMinGasPriceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
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
func (m *QueryMinGasPriceResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMinGasPriceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMinGasPriceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinGasPrice", wireType)
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
			if err := m.MinGasPrice.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryRecommendedGasPriceRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryRecommendedGasPriceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryRecommendedGasPriceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AfterBlocks", wireType)
			}
			m.AfterBlocks = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AfterBlocks |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
func (m *QueryRecommendedGasPriceResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryRecommendedGasPriceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryRecommendedGasPriceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Low", wireType)
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
			if err := m.Low.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Med", wireType)
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
			if err := m.Med.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field High", wireType)
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
			if err := m.High.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
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
func (m *QueryParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
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
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
