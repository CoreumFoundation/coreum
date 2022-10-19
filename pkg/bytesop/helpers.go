package bytesop

import (
	"github.com/gogo/protobuf/proto"
)

func Join(keyComponents ...[]byte) []byte {
	var totalLength int
	for _, v := range keyComponents {
		totalLength += len(v)
	}

	res := make([]byte, 0, totalLength)
	for _, v := range keyComponents {
		res = append(res, v...)
	}
	return res
}

func WithLength(value []byte) []byte {
	lenMarshaled := proto.EncodeVarint(uint64(len(value)))
	res := make([]byte, 0, len(lenMarshaled)+len(value))
	res = append(res, lenMarshaled...)
	return append(res, value...)
}
