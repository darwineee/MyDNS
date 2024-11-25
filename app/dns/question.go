package dns

import (
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/dns/type"
	"encoding/binary"
	"errors"
)

// Question represents a DNS question section
type Question struct {
	Name     *NameAddr
	Type     _type.RecordType
	Class    _type.RecordClass
	calcSize int
}

func parseSingleQuestion(origBuf []byte, startBuf []byte) (*Question, []byte, error) {
	name := &NameAddr{}
	currentBuf := startBuf

	// Parse domain name
	for {
		if len(currentBuf) < 1 {
			return nil, startBuf, errors.New("buffer too small")
		}

		length := currentBuf[0]

		// End of domain name
		if length == 0 {
			name.Encoded = append(name.Encoded, 0)
			currentBuf = currentBuf[1:]
			break
		}

		// Check if it's a pointer (compression)
		if _type.ToLabel(length) == _type.POINTER {
			if len(currentBuf) < 2 {
				return nil, startBuf, errors.New("buffer too small for pointer")
			}

			offset := int(binary.BigEndian.Uint16(currentBuf[:2]) & 0x3FFF)
			if offset >= len(origBuf) {
				return nil, startBuf, errors.New("invalid pointer offset")
			}

			// Store the pointer bytes in encoded
			name.Encoded = append(name.Encoded, currentBuf[0], currentBuf[1])

			// Follow the pointer for string representation
			pointerBuf := origBuf[offset:]
			remainingName, _, err := parseDomainName(origBuf, pointerBuf)
			if err != nil {
				return nil, startBuf, err
			}
			if len(name.String) > 0 {
				name.String += "."
			}
			name.String += remainingName

			currentBuf = currentBuf[2:]
			break
		}

		// Regular label
		if len(currentBuf) < int(length+1) {
			return nil, startBuf, errors.New("buffer too small for label")
		}

		name.Encoded = append(name.Encoded, currentBuf[:length+1]...)
		if len(name.String) > 0 {
			name.String += "."
		}
		name.String += string(currentBuf[1 : length+1])

		currentBuf = currentBuf[length+1:]
	}

	// Parse type and class
	if len(currentBuf) < 4 {
		return nil, startBuf, errors.New("buffer too small for question info")
	}
	qtype := _type.RecordType(binary.BigEndian.Uint16(currentBuf[0:2]))
	qclass := _type.RecordClass(binary.BigEndian.Uint16(currentBuf[2:4]))

	return &Question{
		Name:  name,
		Type:  qtype,
		Class: qclass,
	}, currentBuf[4:], nil
}

func parseDomainName(origBuf []byte, startBuf []byte) (string, []byte, error) {
	var name string
	currentBuf := startBuf

	for {
		if len(currentBuf) < 1 {
			return "", startBuf, errors.New("buffer too small")
		}

		length := currentBuf[0]

		if length == 0 {
			return name, currentBuf[1:], nil
		}

		if _type.ToLabel(length) == _type.POINTER {
			if len(currentBuf) < 2 {
				return "", startBuf, errors.New("buffer too small for pointer")
			}

			offset := int(binary.BigEndian.Uint16(currentBuf[:2]) & 0x3FFF)
			if offset >= len(origBuf) {
				return "", startBuf, errors.New("invalid pointer offset")
			}

			pointerBuf := origBuf[offset:]
			remainingName, _, err := parseDomainName(origBuf, pointerBuf)
			if err != nil {
				return "", startBuf, err
			}
			if len(name) > 0 {
				name += "."
			}
			name += remainingName
			return name, currentBuf[2:], nil
		}

		if len(currentBuf) < int(length+1) {
			return "", startBuf, errors.New("buffer too small for label")
		}

		if len(name) > 0 {
			name += "."
		}
		name += string(currentBuf[1 : length+1])

		currentBuf = currentBuf[length+1:]
	}
}

func parseQuestions(buf []byte, count uint16) ([]*Question, error) {
	questions := make([]*Question, 0, count)
	currentBuf := buf

	for i := uint16(0); i < count; i++ {
		if len(currentBuf) == 0 {
			return nil, errors.New("buffer exhausted before parsing all questions")
		}
		if len(currentBuf) >= config.PkgLimitRFC1035 {
			return nil, errors.New("buffer exceeds UDP packet limit")
		}

		question, remaining, err := parseSingleQuestion(buf, currentBuf)
		if err != nil {
			return nil, err
		}

		questions = append(questions, question)
		currentBuf = remaining
	}

	return questions, nil
}

// Size in byte
func (q *Question) Size() int {
	if q.calcSize == 0 {
		q.calcSize = 4 + q.Name.Size()
	}
	return q.calcSize
}

// WriteTo a byte buffer with content of a DNS Message
func (q *Question) WriteTo(buf []byte) ([]byte, error) {
	if len(buf) < q.Size() {
		return buf, errors.New("buffer too small")
	}
	remaining, err := q.Name.WriteTo(buf)
	if err != nil {
		return buf, err
	}

	binary.BigEndian.PutUint16(remaining, uint16(q.Type))
	binary.BigEndian.PutUint16(remaining[2:], uint16(q.Class))

	return remaining[4:], nil
}
