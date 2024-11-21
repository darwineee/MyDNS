package dns

import (
	"encoding/binary"
	"errors"
)

// Header represents the structure of a DNS message Header
type Header struct {
	ID                  uint16 // 16 bits: Packet Identifier
	QueryResponse       bool   // 1 bit: 1 for response, 0 for query
	OperationCode       uint8  // 4 bits: Kind of query
	AuthoritativeAnswer bool   // 1 bit: 1 if server is authoritative
	Truncation          bool   // 1 bit: 1 if message is larger than 512 bytes
	RecursionDesired    bool   // 1 bit: 1 if recursion is desired
	RecursionAvailable  bool   // 1 bit: 1 if recursion is available
	Reserved            uint8  // 3 bits: Reserved for future use (DNSSEC)
	ResponseCode        uint8  // 4 bits: Response code
	QuestionCount       uint16 // 16 bits: Number of Questions
	AnswerCount         uint16 // 16 bits: Number of Answers
	AuthorityCount      uint16 // 16 bits: Number of authority records
	AdditionalCount     uint16 // 16 bits: Number of additional records
}

func (h *Header) Size() int {
	return 12
}

func parseHeader(buf []byte) (*Header, []byte, error) {
	if len(buf) < 12 {
		return nil, buf, errors.New("buffer too small for Header")
	}

	h := &Header{
		ID: binary.BigEndian.Uint16(buf[0:2]),
	}

	// Parse flags
	flags := binary.BigEndian.Uint16(buf[2:4])
	h.QueryResponse = (flags & 0x8000) != 0
	h.OperationCode = uint8((flags & 0x7800) >> 11)
	h.AuthoritativeAnswer = (flags & 0x0400) != 0
	h.Truncation = (flags & 0x0200) != 0
	h.RecursionDesired = (flags & 0x0100) != 0
	h.RecursionAvailable = (flags & 0x0080) != 0
	h.Reserved = uint8((flags & 0x0070) >> 4)
	h.ResponseCode = uint8(flags & 0x000F)

	// Parse counts
	h.QuestionCount = binary.BigEndian.Uint16(buf[4:6])
	h.AnswerCount = binary.BigEndian.Uint16(buf[6:8])
	h.AuthorityCount = binary.BigEndian.Uint16(buf[8:10])
	h.AdditionalCount = binary.BigEndian.Uint16(buf[10:12])

	return h, buf[12:], nil
}

// WriteTo a byte buffer with content of a DNS Message
func (h *Header) WriteTo(buf []byte) ([]byte, error) {
	if len(buf) < h.Size() {
		return buf, errors.New("buffer too small")
	}

	// Write ID
	binary.BigEndian.PutUint16(buf, h.ID)

	// Construct flags
	var flags uint16
	if h.QueryResponse {
		flags |= 0x8000
	}
	flags |= uint16(h.OperationCode) << 11
	if h.AuthoritativeAnswer {
		flags |= 0x0400
	}
	if h.Truncation {
		flags |= 0x0200
	}
	if h.RecursionDesired {
		flags |= 0x0100
	}
	if h.RecursionAvailable {
		flags |= 0x0080
	}
	flags |= uint16(h.Reserved) << 4
	flags |= uint16(h.ResponseCode)

	// Write flags
	binary.BigEndian.PutUint16(buf[2:], flags)

	// Write counts
	binary.BigEndian.PutUint16(buf[4:], h.QuestionCount)
	binary.BigEndian.PutUint16(buf[6:], h.AnswerCount)
	binary.BigEndian.PutUint16(buf[8:], h.AuthorityCount)
	binary.BigEndian.PutUint16(buf[10:], h.AdditionalCount)

	// Return the remaining buffer
	return buf[12:], nil
}
