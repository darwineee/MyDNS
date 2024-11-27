package dns

import (
	"errors"
)

type Addr struct {
	Encoded []byte
	String  string
}

// Size in byte
func (addr *Addr) Size() int {
	return len(addr.Encoded)
}

// WriteTo a byte buffer with content of a DNS Message
func (addr *Addr) WriteTo(buf []byte) ([]byte, error) {
	if len(buf) < addr.Size() {
		return buf, errors.New("buffer too small for name")
	}
	copy(buf, addr.Encoded)
	return buf[addr.Size():], nil
}
