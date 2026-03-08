// Package identity provides domain-layer ID generation without external dependencies.
// Domain layers must have ZERO external deps; this package uses crypto/rand
// from the standard library to generate unique identifiers.
package identity

import (
	"crypto/rand"
	"fmt"
)

// NewID generates a random UUID v4 string using only the standard library.
func NewID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
