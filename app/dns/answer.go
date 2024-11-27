package dns

import (
	_type "com.sentry.dev/app/dns/type"
	"encoding/binary"
	"errors"
	"net"
)

type Answer struct {
	Pointer  uint16
	Type     _type.RecordType
	Class    _type.RecordClass
	TTL      uint32
	RDLength uint16
	RData    net.IP
	calcSize int
}

// Size in byte
func (a *Answer) Size() int {
	if a.calcSize == 0 {
		a.calcSize = 12 + len(a.RData)
	}
	return a.calcSize
}

// WriteTo a byte buffer with content of a DNS Message
func (a *Answer) WriteTo(buf []byte) ([]byte, error) {
	if len(buf) < a.Size() {
		return buf, errors.New("buffer too small for answer")
	}

	binary.BigEndian.PutUint16(buf, a.Pointer)
	binary.BigEndian.PutUint16(buf[2:], uint16(a.Type))
	binary.BigEndian.PutUint16(buf[4:], uint16(a.Class))
	binary.BigEndian.PutUint32(buf[6:], a.TTL)
	binary.BigEndian.PutUint16(buf[10:], a.RDLength)

	copy(buf[12:], a.RData)
	return buf[12+len(a.RData):], nil
}
