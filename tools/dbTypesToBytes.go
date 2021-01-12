package tools

func GetPaddedString(s string) []byte {
	l := (len(s) + 1) % 8
	if l != 0 {
		l = 8 - l
	}
	a := make([]byte, l)
	for i := range a {
		a[i] = []byte("X")[0]
	}
	return append(GetString(s), a...)
}

func GetString(s string) []byte {
	return []byte(s + "\x00")
}

func NumBytes(n uint32, bigEndian bool) []byte {
	if bigEndian {
		return []byte{byte((n >> (8 * 3)) & 0xFF), byte((n >> (8 * 2)) & 0xFF), byte((n >> (8)) & 0xFF), byte(n & 0xFF)}[:4]
	}
	return []byte{byte(n & 0xFF), byte((n >> (8)) & 0xFF), byte((n >> (8 * 2)) & 0xFF), byte((n >> (8 * 3)) & 0xFF)}[:4]
}

func MinBytes() []byte {
	return []byte{0x00, 0x00, 0x00, 0x00}
}

func MaxBytes() []byte {
	return []byte{0xFF, 0xFF, 0xFF, 0xFF}
}

func BytesNum(bytes []byte, bigEndian bool) int32 {
	if bigEndian {
		return int32(bytes[3]) | int32(bytes[2])<<8 | int32(bytes[1])<<(8*2) | int32(bytes[0])<<(8*3)
	}
	return int32(bytes[3])<<(8*3) | int32(bytes[2])<<(8*2) | int32(bytes[1])<<8 | int32(bytes[0])
}
