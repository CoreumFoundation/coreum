// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: coreum/feemodel/v1/params.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
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

// ModelParams define fee model params.
// There are four regions on the fee model curve
// - between 0 and "long average block gas" where gas price goes down exponentially from InitialGasPrice to gas price with maximum discount (InitialGasPrice * (1 - MaxDiscount))
// - between "long average block gas" and EscalationStartBlockGas (EscalationStartBlockGas = MaxBlockGas * EscalationStartFraction) where we offer gas price with maximum discount all the time
// - between EscalationStartBlockGas (EscalationStartBlockGas = MaxBlockGas * EscalationStartFraction) and MaxBlockGas where price goes up rapidly (being an output of a power function) from gas price with maximum discount to MaxGasPrice  (MaxGasPrice = InitialGasPrice * MaxGasMultiplier)
// - above MaxBlockGas (if it happens for any reason) where price is equal to MaxGasPrice (MaxGasPrice = InitialGasPrice * MaxGasMultiplier)
//
// The input (x value) for that function is calculated by taking short block gas average.
// Price (y value) being an output of the fee model is used as the minimum gas price for next block.
type ModelParams struct {
	// initial_gas_price is used when block gas short average is 0. It happens when there are no transactions being broadcasted. This value is also used to initialize gas price on brand-new chain.
	InitialGasPrice github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=initial_gas_price,json=initialGasPrice,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"initial_gas_price" yaml:"initial_gas_price"`
	// max_gas_price_multiplier is used to compute max_gas_price (max_gas_price = initial_gas_price * max_gas_price_multiplier). Max gas price is charged when block gas short average is greater than or equal to MaxBlockGas. This value is used to limit gas price escalation to avoid having possible infinity GasPrice value otherwise.
	MaxGasPriceMultiplier github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=max_gas_price_multiplier,json=maxGasPriceMultiplier,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"max_gas_price_multiplier" yaml:"max_gas_price_multiplier"`
	// max_discount is th maximum discount we offer on top of initial gas price if short average block gas is between long average block gas and escalation start block gas.
	MaxDiscount github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,opt,name=max_discount,json=maxDiscount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"max_discount" yaml:"max_discount"`
	// escalation_start_fraction defines fraction of max block gas usage where gas price escalation starts if short average block gas is higher than this value.
	EscalationStartFraction github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,4,opt,name=escalation_start_fraction,json=escalationStartFraction,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"escalation_start_fraction" yaml:"escalation_start_fraction"`
	// max_block_gas sets the maximum capacity of block. This is enforced on tendermint level in genesis configuration. Once short average block gas goes above this value, gas price is a flat line equal to MaxGasPrice.
	MaxBlockGas int64 `protobuf:"varint,5,opt,name=max_block_gas,json=maxBlockGas,proto3" json:"max_block_gas,omitempty" yaml:"max_block_gas"`
	// short_ema_block_length defines inertia for short average long gas in EMA model. The equation is: NewAverage = ((ShortAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / ShortAverageBlockLength
	// The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation.
	ShortEmaBlockLength uint32 `protobuf:"varint,6,opt,name=short_ema_block_length,json=shortEmaBlockLength,proto3" json:"short_ema_block_length,omitempty" yaml:"short_ema_block_length"`
	// long_ema_block_length defines inertia for long average block gas in EMA model. The equation is: NewAverage = ((LongAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / LongAverageBlockLength
	// The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation.
	LongEmaBlockLength uint32 `protobuf:"varint,7,opt,name=long_ema_block_length,json=longEmaBlockLength,proto3" json:"long_ema_block_length,omitempty" yaml:"long_ema_block_length"`
}

func (m *ModelParams) Reset()         { *m = ModelParams{} }
func (m *ModelParams) String() string { return proto.CompactTextString(m) }
func (*ModelParams) ProtoMessage()    {}
func (*ModelParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_3500559e6fedefd6, []int{0}
}
func (m *ModelParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ModelParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ModelParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ModelParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ModelParams.Merge(m, src)
}
func (m *ModelParams) XXX_Size() int {
	return m.Size()
}
func (m *ModelParams) XXX_DiscardUnknown() {
	xxx_messageInfo_ModelParams.DiscardUnknown(m)
}

var xxx_messageInfo_ModelParams proto.InternalMessageInfo

func (m *ModelParams) GetMaxBlockGas() int64 {
	if m != nil {
		return m.MaxBlockGas
	}
	return 0
}

func (m *ModelParams) GetShortEmaBlockLength() uint32 {
	if m != nil {
		return m.ShortEmaBlockLength
	}
	return 0
}

func (m *ModelParams) GetLongEmaBlockLength() uint32 {
	if m != nil {
		return m.LongEmaBlockLength
	}
	return 0
}

// Params store gov manageable feemodel parameters.
type Params struct {
	// model is a fee model params.
	Model ModelParams `protobuf:"bytes,1,opt,name=model,proto3" json:"model" yaml:"model"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_3500559e6fedefd6, []int{1}
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

func (m *Params) GetModel() ModelParams {
	if m != nil {
		return m.Model
	}
	return ModelParams{}
}

func init() {
	proto.RegisterType((*ModelParams)(nil), "coreum.feemodel.v1.ModelParams")
	proto.RegisterType((*Params)(nil), "coreum.feemodel.v1.Params")
}

func init() { proto.RegisterFile("coreum/feemodel/v1/params.proto", fileDescriptor_3500559e6fedefd6) }

var fileDescriptor_3500559e6fedefd6 = []byte{
	// 510 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x4f, 0x6f, 0xd3, 0x30,
	0x1c, 0x6d, 0x18, 0x2d, 0xc2, 0xdd, 0x84, 0xf0, 0x3a, 0x08, 0x08, 0xe2, 0xe2, 0x03, 0xea, 0x85,
	0x44, 0x63, 0x37, 0xc4, 0x29, 0xec, 0x8f, 0x04, 0x54, 0x1a, 0x99, 0xe0, 0xc0, 0x25, 0x72, 0x53,
	0x2f, 0x8d, 0x16, 0xc7, 0x51, 0xec, 0x54, 0xdd, 0x57, 0xe0, 0x80, 0xf8, 0x58, 0x3b, 0xee, 0x88,
	0x38, 0x44, 0xa8, 0xfd, 0x06, 0x39, 0x71, 0x44, 0x76, 0xd2, 0x65, 0x53, 0xb7, 0x43, 0x4f, 0xc9,
	0xef, 0xf7, 0x9e, 0xdf, 0xb3, 0xfd, 0xf3, 0x03, 0x28, 0xe0, 0x19, 0xcd, 0x99, 0x73, 0x4a, 0x29,
	0xe3, 0x63, 0x1a, 0x3b, 0xd3, 0x5d, 0x27, 0x25, 0x19, 0x61, 0xc2, 0x4e, 0x33, 0x2e, 0x39, 0x84,
	0x15, 0xc1, 0x5e, 0x12, 0xec, 0xe9, 0xee, 0xf3, 0x5e, 0xc8, 0x43, 0xae, 0x61, 0x47, 0xfd, 0x55,
	0x4c, 0xfc, 0xaf, 0x0d, 0xba, 0x43, 0x45, 0x39, 0xd6, 0xeb, 0xe1, 0x14, 0x3c, 0x8e, 0x92, 0x48,
	0x46, 0x24, 0xf6, 0x43, 0x22, 0xfc, 0x34, 0x8b, 0x02, 0x6a, 0x1a, 0x7d, 0x63, 0xf0, 0xd0, 0xfd,
	0x78, 0x51, 0xa0, 0xd6, 0x9f, 0x02, 0xbd, 0x0e, 0x23, 0x39, 0xc9, 0x47, 0x76, 0xc0, 0x99, 0x13,
	0x70, 0xc1, 0xb8, 0xa8, 0x3f, 0x6f, 0xc4, 0xf8, 0xcc, 0x91, 0xe7, 0x29, 0x15, 0xf6, 0x3e, 0x0d,
	0xca, 0x02, 0x99, 0xe7, 0x84, 0xc5, 0xef, 0xf0, 0x8a, 0x20, 0xf6, 0x1e, 0xd5, 0xbd, 0x23, 0x22,
	0x8e, 0x55, 0x07, 0xfe, 0x30, 0x80, 0xc9, 0xc8, 0xac, 0xe1, 0xf8, 0x2c, 0x8f, 0x65, 0x94, 0xc6,
	0x11, 0xcd, 0xcc, 0x7b, 0xda, 0xff, 0xcb, 0xda, 0xfe, 0xa8, 0xf2, 0xbf, 0x4b, 0x17, 0x7b, 0x3b,
	0x8c, 0xcc, 0x96, 0x5b, 0x18, 0x5e, 0xf5, 0xe1, 0x04, 0x6c, 0xaa, 0x35, 0xe3, 0x48, 0x04, 0x3c,
	0x4f, 0xa4, 0xb9, 0xa1, 0xfd, 0x0f, 0xd6, 0xf6, 0xdf, 0x6e, 0xfc, 0x97, 0x5a, 0xd8, 0xeb, 0x32,
	0x32, 0xdb, 0xaf, 0x2b, 0xf8, 0xd3, 0x00, 0xcf, 0xa8, 0x08, 0x48, 0x4c, 0x64, 0xc4, 0x13, 0x5f,
	0x48, 0x92, 0x49, 0xff, 0x34, 0x23, 0x81, 0x2a, 0xcd, 0xfb, 0xda, 0xd7, 0x5b, 0xdb, 0xb7, 0x5f,
	0xf9, 0xde, 0x29, 0x8c, 0xbd, 0xa7, 0x0d, 0x76, 0xa2, 0xa0, 0xc3, 0x1a, 0x81, 0xef, 0xc1, 0x96,
	0xda, 0xee, 0x28, 0xe6, 0xc1, 0x99, 0xba, 0x34, 0xb3, 0xdd, 0x37, 0x06, 0x1b, 0xae, 0x59, 0x16,
	0xa8, 0xd7, 0x9c, 0xe6, 0x0a, 0xae, 0x8e, 0xe3, 0xaa, 0xf2, 0x88, 0x08, 0xf8, 0x0d, 0x3c, 0x11,
	0x13, 0x9e, 0x49, 0x9f, 0x32, 0x52, 0x93, 0x62, 0x9a, 0x84, 0x72, 0x62, 0x76, 0xfa, 0xc6, 0x60,
	0xcb, 0x7d, 0x55, 0x16, 0xe8, 0x65, 0x25, 0x73, 0x3b, 0x0f, 0x7b, 0xdb, 0x1a, 0x38, 0x60, 0x44,
	0x8b, 0x7e, 0xd6, 0x5d, 0x78, 0x02, 0x76, 0x62, 0x9e, 0x84, 0xab, 0xb2, 0x0f, 0xb4, 0x6c, 0xbf,
	0x2c, 0xd0, 0x8b, 0x4a, 0xf6, 0x56, 0x1a, 0xf6, 0xa0, 0xea, 0xdf, 0x14, 0xc5, 0x5f, 0x41, 0xa7,
	0x7e, 0xf4, 0x9f, 0x40, 0x5b, 0xc7, 0x44, 0x3f, 0xf4, 0xee, 0x5b, 0x64, 0xaf, 0xc6, 0xc7, 0xbe,
	0x16, 0x12, 0xb7, 0xa7, 0x26, 0x52, 0x16, 0x68, 0xb3, 0xbe, 0x11, 0x05, 0x61, 0xaf, 0xd2, 0x70,
	0x87, 0x17, 0x73, 0xcb, 0xb8, 0x9c, 0x5b, 0xc6, 0xdf, 0xb9, 0x65, 0xfc, 0x5a, 0x58, 0xad, 0xcb,
	0x85, 0xd5, 0xfa, 0xbd, 0xb0, 0x5a, 0xdf, 0xf7, 0xae, 0x0d, 0xf0, 0x83, 0x76, 0x38, 0xe4, 0x79,
	0x32, 0xd6, 0x53, 0x70, 0xea, 0x48, 0xcf, 0x9a, 0x50, 0xeb, 0x89, 0x8e, 0x3a, 0x3a, 0xa7, 0x7b,
	0xff, 0x03, 0x00, 0x00, 0xff, 0xff, 0x32, 0xfa, 0xeb, 0xef, 0xf4, 0x03, 0x00, 0x00,
}

func (m *ModelParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ModelParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ModelParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.LongEmaBlockLength != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.LongEmaBlockLength))
		i--
		dAtA[i] = 0x38
	}
	if m.ShortEmaBlockLength != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.ShortEmaBlockLength))
		i--
		dAtA[i] = 0x30
	}
	if m.MaxBlockGas != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxBlockGas))
		i--
		dAtA[i] = 0x28
	}
	{
		size := m.EscalationStartFraction.Size()
		i -= size
		if _, err := m.EscalationStartFraction.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	{
		size := m.MaxDiscount.Size()
		i -= size
		if _, err := m.MaxDiscount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size := m.MaxGasPriceMultiplier.Size()
		i -= size
		if _, err := m.MaxGasPriceMultiplier.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.InitialGasPrice.Size()
		i -= size
		if _, err := m.InitialGasPrice.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
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
		size, err := m.Model.MarshalToSizedBuffer(dAtA[:i])
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
func (m *ModelParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.InitialGasPrice.Size()
	n += 1 + l + sovParams(uint64(l))
	l = m.MaxGasPriceMultiplier.Size()
	n += 1 + l + sovParams(uint64(l))
	l = m.MaxDiscount.Size()
	n += 1 + l + sovParams(uint64(l))
	l = m.EscalationStartFraction.Size()
	n += 1 + l + sovParams(uint64(l))
	if m.MaxBlockGas != 0 {
		n += 1 + sovParams(uint64(m.MaxBlockGas))
	}
	if m.ShortEmaBlockLength != 0 {
		n += 1 + sovParams(uint64(m.ShortEmaBlockLength))
	}
	if m.LongEmaBlockLength != 0 {
		n += 1 + sovParams(uint64(m.LongEmaBlockLength))
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Model.Size()
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ModelParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ModelParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ModelParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InitialGasPrice", wireType)
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
			if err := m.InitialGasPrice.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxGasPriceMultiplier", wireType)
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
			if err := m.MaxGasPriceMultiplier.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxDiscount", wireType)
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
			if err := m.MaxDiscount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EscalationStartFraction", wireType)
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
			if err := m.EscalationStartFraction.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxBlockGas", wireType)
			}
			m.MaxBlockGas = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxBlockGas |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ShortEmaBlockLength", wireType)
			}
			m.ShortEmaBlockLength = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ShortEmaBlockLength |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LongEmaBlockLength", wireType)
			}
			m.LongEmaBlockLength = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LongEmaBlockLength |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
				return fmt.Errorf("proto: wrong wireType = %d for field Model", wireType)
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
			if err := m.Model.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
