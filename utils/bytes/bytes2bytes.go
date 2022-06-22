package bytes

func BytesTo32Bytes(bytes []byte) [32]byte {
	var res [32]byte
	copy(res[:], bytes)

	return res
}

func BytesTo65Bytes(bytes []byte) [65]byte {
	var res [65]byte
	copy(res[:], bytes)

	return res
}
