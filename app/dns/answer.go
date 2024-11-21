package dns

import (
	_type "com.sentry.dev/app/dns/type"
	"encoding/binary"
	"errors"
)

type Answer struct {
	Name     *NameAddr
	Type     _type.RecordType
	Class    _type.RecordClass
	TTL      uint32
	RDLength uint16
	RData    *IPAddr
}

// Size in byte
func (a *Answer) Size() int {
	return 10 + a.Name.Size() + a.RData.Size()
}

// WriteTo a byte buffer with content of a DNS Message
func (a *Answer) WriteTo(buf []byte) ([]byte, error) {
	if len(buf) < a.Size() {
		return buf, errors.New("buffer too small for answer")
	}

	remaining, err := a.Name.WriteTo(buf)
	if err != nil {
		return buf, err
	}

	binary.BigEndian.PutUint16(remaining, uint16(a.Type))
	binary.BigEndian.PutUint16(remaining[2:], uint16(a.Class))
	binary.BigEndian.PutUint32(remaining[4:], a.TTL)
	binary.BigEndian.PutUint16(remaining[8:], a.RDLength)

	remaining, err = a.RData.WriteTo(remaining[10:])
	if err != nil {
		return buf, err
	}

	return remaining, nil
}
