package database

const (
	BigEndian    = Endian(true)
	LittleEndian = Endian(false)
)

// Endian represents how the media player's processor stores numbers.
type Endian bool

// NumBytes converts an int32 to a slice of bytes.
// The length of this slice should be exactly four.
func (e Endian) NumBytes(n int32) []byte {
	if e == BigEndian {
		return []byte{byte((n >> (8 * 3)) & 0xFF), byte((n >> (8 * 2)) & 0xFF), byte((n >> (8)) & 0xFF), byte(n & 0xFF)}[:4]
	}
	return []byte{byte(n & 0xFF), byte((n >> (8)) & 0xFF), byte((n >> (8 * 2)) & 0xFF), byte((n >> (8 * 3)) & 0xFF)}[:4]
}

// BytesNum converts a slice of bytes to an int32.
// Only the first four items of the slice are used.
//
// Bug(GYBATTF): Does not handle slices that have a length < 4.
func (e Endian) BytesNum(bytes []byte) int32 {
	if e == BigEndian {
		return int32(bytes[3]) | int32(bytes[2])<<8 | int32(bytes[1])<<(8*2) | int32(bytes[0])<<(8*3)
	}
	return int32(bytes[3])<<(8*3) | int32(bytes[2])<<(8*2) | int32(bytes[1])<<8 | int32(bytes[0])
}
