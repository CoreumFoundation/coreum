package types

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
