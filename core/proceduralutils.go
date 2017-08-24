package core

import "hash/crc32"

// HashU32 computes an unsigned 32 bit hash value for the given parameter
func HashU32(x uint32) uint32 {
	if x == 0 {
		panic("Cannot hash zero, this should __never__ happen")
	}

	x += x << 10
	x ^= x >> 6
	x += x << 3
	x ^= x >> 11
	x += x << 15
	return x
}

// HashStringToU32 computes an unsigned 32 bit hash value for the given parameter
func HashStringToU32(x string) uint32 {
	h := crc32.NewIEEE()
	if _, err := h.Write([]byte(x)); err != nil {
		panic("Error hashing value. This should never happen.")
	}
	v := h.Sum32()
	return HashU32(v)
}
