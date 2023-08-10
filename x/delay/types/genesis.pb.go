// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/delay/v1/genesis.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
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

// GenesisState defines the module genesis state.
type GenesisState struct {
	// tokens keep the fungible token state
	DelayedItems []DelayedItem `protobuf:"bytes,1,rep,name=delayed_items,json=delayedItems,proto3" json:"delayed_items"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_97754df78b5c97b3, []int{0}
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

func (m *GenesisState) GetDelayedItems() []DelayedItem {
	if m != nil {
		return m.DelayedItems
	}
	return nil
}

type DelayedItem struct {
	Id            string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ExecutionTime time.Time  `protobuf:"bytes,2,opt,name=execution_time,json=executionTime,proto3,stdtime" json:"execution_time"`
	Data          *types.Any `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *DelayedItem) Reset()         { *m = DelayedItem{} }
func (m *DelayedItem) String() string { return proto.CompactTextString(m) }
func (*DelayedItem) ProtoMessage()    {}
func (*DelayedItem) Descriptor() ([]byte, []int) {
	return fileDescriptor_97754df78b5c97b3, []int{1}
}
func (m *DelayedItem) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DelayedItem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DelayedItem.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DelayedItem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DelayedItem.Merge(m, src)
}
func (m *DelayedItem) XXX_Size() int {
	return m.Size()
}
func (m *DelayedItem) XXX_DiscardUnknown() {
	xxx_messageInfo_DelayedItem.DiscardUnknown(m)
}

var xxx_messageInfo_DelayedItem proto.InternalMessageInfo

func (m *DelayedItem) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *DelayedItem) GetExecutionTime() time.Time {
	if m != nil {
		return m.ExecutionTime
	}
	return time.Time{}
}

func (m *DelayedItem) GetData() *types.Any {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "coreum.delay.v1.GenesisState")
	proto.RegisterType((*DelayedItem)(nil), "coreum.delay.v1.DelayedItem")
}

func init() { proto.RegisterFile("coreum/delay/v1/genesis.proto", fileDescriptor_97754df78b5c97b3) }

var fileDescriptor_97754df78b5c97b3 = []byte{
	// 328 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0xcd, 0x4e, 0x02, 0x31,
	0x14, 0x85, 0xa7, 0x40, 0x8c, 0x96, 0x1f, 0x93, 0x09, 0x8b, 0x91, 0xe8, 0x40, 0x58, 0xcd, 0xaa,
	0x0d, 0xf0, 0x04, 0xa2, 0x91, 0x18, 0xe3, 0x66, 0x34, 0x31, 0x71, 0x43, 0x0a, 0xbd, 0x8e, 0x4d,
	0x98, 0x29, 0xa1, 0x1d, 0xc2, 0xbc, 0x05, 0x0b, 0x1f, 0x8a, 0x25, 0x4b, 0x57, 0x6a, 0xe0, 0x45,
	0xcc, 0xb4, 0xa0, 0x06, 0x77, 0x6d, 0xcf, 0xd7, 0x7b, 0xce, 0xcd, 0xc1, 0x17, 0x63, 0x39, 0x83,
	0x34, 0xa6, 0x1c, 0x26, 0x2c, 0xa3, 0xf3, 0x0e, 0x8d, 0x20, 0x01, 0x25, 0x14, 0x99, 0xce, 0xa4,
	0x96, 0xee, 0xa9, 0x95, 0x89, 0x91, 0xc9, 0xbc, 0xd3, 0xa8, 0x47, 0x32, 0x92, 0x46, 0xa3, 0xf9,
	0xc9, 0x62, 0x8d, 0x66, 0x24, 0x65, 0x34, 0x01, 0x6a, 0x6e, 0xa3, 0xf4, 0x85, 0x6a, 0x11, 0x83,
	0xd2, 0x2c, 0x9e, 0xee, 0x80, 0xb3, 0x43, 0x80, 0x25, 0x99, 0x95, 0xda, 0x4f, 0xb8, 0x32, 0xb0,
	0x9e, 0x0f, 0x9a, 0x69, 0x70, 0x07, 0xb8, 0x6a, 0xdc, 0x80, 0x0f, 0x85, 0x86, 0x58, 0x79, 0xa8,
	0x55, 0x0c, 0xca, 0xdd, 0x73, 0x72, 0x10, 0x85, 0x5c, 0x5b, 0xea, 0x56, 0x43, 0xdc, 0x2f, 0xad,
	0x3e, 0x9a, 0x4e, 0x58, 0xe1, 0xbf, 0x4f, 0xaa, 0xfd, 0x86, 0x70, 0xf9, 0x0f, 0xe3, 0xd6, 0x70,
	0x41, 0x70, 0x0f, 0xb5, 0x50, 0x70, 0x12, 0x16, 0x04, 0x77, 0xef, 0x70, 0x0d, 0x16, 0x30, 0x4e,
	0xb5, 0x90, 0xc9, 0x30, 0x0f, 0xec, 0x15, 0x5a, 0x28, 0x28, 0x77, 0x1b, 0xc4, 0x86, 0x25, 0xfb,
	0xb0, 0xe4, 0x71, 0xbf, 0x4d, 0xff, 0x38, 0xf7, 0x59, 0x7e, 0x36, 0x51, 0x58, 0xfd, 0xf9, 0x9b,
	0xab, 0x6e, 0x80, 0x4b, 0x9c, 0x69, 0xe6, 0x15, 0xcd, 0x88, 0xfa, 0xbf, 0x11, 0x97, 0x49, 0x16,
	0x1a, 0xa2, 0x7f, 0xbf, 0xda, 0xf8, 0x68, 0xbd, 0xf1, 0xd1, 0xd7, 0xc6, 0x47, 0xcb, 0xad, 0xef,
	0xac, 0xb7, 0xbe, 0xf3, 0xbe, 0xf5, 0x9d, 0xe7, 0x5e, 0x24, 0xf4, 0x6b, 0x3a, 0x22, 0x63, 0x19,
	0xd3, 0x2b, 0xb3, 0xec, 0x8d, 0x4c, 0x13, 0xce, 0x72, 0x13, 0xba, 0xeb, 0x69, 0xde, 0xa5, 0x8b,
	0x5d, 0x59, 0x3a, 0x9b, 0x82, 0x1a, 0x1d, 0x19, 0x8b, 0xde, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x4e, 0x9d, 0x1c, 0x8d, 0xc9, 0x01, 0x00, 0x00,
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
	if len(m.DelayedItems) > 0 {
		for iNdEx := len(m.DelayedItems) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DelayedItems[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *DelayedItem) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DelayedItem) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DelayedItem) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Data != nil {
		{
			size, err := m.Data.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.ExecutionTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.ExecutionTime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintGenesis(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Id)))
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
	if len(m.DelayedItems) > 0 {
		for _, e := range m.DelayedItems {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *DelayedItem) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.ExecutionTime)
	n += 1 + l + sovGenesis(uint64(l))
	if m.Data != nil {
		l = m.Data.Size()
		n += 1 + l + sovGenesis(uint64(l))
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
				return fmt.Errorf("proto: wrong wireType = %d for field DelayedItems", wireType)
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
			m.DelayedItems = append(m.DelayedItems, DelayedItem{})
			if err := m.DelayedItems[len(m.DelayedItems)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *DelayedItem) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: DelayedItem: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DelayedItem: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
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
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExecutionTime", wireType)
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
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.ExecutionTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
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
			if m.Data == nil {
				m.Data = &types.Any{}
			}
			if err := m.Data.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
