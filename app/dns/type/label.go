package _type

// Label represents the type of DNS label octet
type Label int

const (
	POINTER Label = iota // Compression pointer (0xC0)
	LEN                  // Length label
)

// ToLabel determines the type of DNS label octet
func ToLabel(val byte) Label {
	if val&0xC0 == 0xC0 { // Check if two highest bits are set (11xxxxxx)
		return POINTER
	}
	return LEN
}

func ToPointer(val uint16) uint16 {
	return val | 0xC000 // 0xC000 = 1100 0000 0000 0000
}
