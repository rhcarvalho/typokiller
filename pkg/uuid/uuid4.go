package uuid

import (
	"crypto/rand"
	"encoding/hex"
)

// UUID4 generates a random UUID (version 4) according to RFC 4122.
func UUID4() []byte {
	// Read 16 random bytes.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// Generate UUID and return formatted fields.
	return format(uuid4(b))
}

// uuid4 changes in place a slice of bytes to make it a UUID (version 4)
// according to RFC 4122.
func uuid4(b []byte) []byte {
	// Set the variant to RFC 4122.
	b[8] &^= 0xc0
	b[8] |= 0x80
	// Set the version number.
	b[6] &^= 0xf0
	b[6] |= 0x40
	return b
}

// format encodes an UUID as hexadecimal with field separators according to RFC
// 4122.
func format(b []byte) []byte {
	// Encode as hexadecimal.
	h := make([]byte, 36)
	hex.Encode(h, b)
	// Shift UUID fields and insert field separators.
	copy(h[24:36], h[20:32])
	copy(h[19:23], h[16:20])
	copy(h[14:18], h[12:16])
	copy(h[9:13], h[8:12])
	h[23] = '-'
	h[18] = '-'
	h[13] = '-'
	h[8] = '-'
	return h
}
