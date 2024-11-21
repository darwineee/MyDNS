package dns

import (
	"errors"
)

type Message struct {
	Header    *Header
	Questions []*Question
	Answers   []*Answer
}

// Size in byte
func (m *Message) Size() int {
	size := m.Header.Size()
	for _, q := range m.Questions {
		size += q.Size()
	}
	for _, a := range m.Answers {
		size += a.Size()
	}
	return size
}

// ParseMessage parses a complete DNS message
func ParseMessage(buf []byte) (*Header, []*Question, error) {
	header, remaining, err := parseHeader(buf)
	if err != nil {
		return nil, nil, err
	}

	questions, err := parseQuestions(remaining, header.QuestionCount)
	if err != nil {
		return header, nil, err
	}

	return header, questions, nil
}

// WriteTo a byte buffer with content of a DNS Message
func (m *Message) WriteTo(buf []byte) (remaining []byte, err error) {
	if len(buf) < m.Size() {
		return buf, errors.New("buffer Size too small")
	}
	remaining, err = m.Header.WriteTo(buf)
	if err != nil {
		return remaining, err
	}
	for _, q := range m.Questions {
		remaining, err = q.WriteTo(remaining)
		if err != nil {
			return remaining, err
		}
	}
	for _, a := range m.Answers {
		remaining, err = a.WriteTo(remaining)
		if err != nil {
			return remaining, err
		}
	}
	return remaining, nil
}
