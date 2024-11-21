package _type

// RecordType represents DNS record types
type RecordType uint16

const (
	TypeA     RecordType = 1  // Host address
	TypeNS    RecordType = 2  // Authoritative name server
	TypeCNAME RecordType = 5  // Canonical name for an alias
	TypeSOA   RecordType = 6  // Start of zone of authority
	TypeWKS   RecordType = 11 // Well known service description
	TypePTR   RecordType = 12 // Domain name pointer
	TypeHINFO RecordType = 13 // Host information
	TypeMINFO RecordType = 14 // Mailbox or mail list information
	TypeMX    RecordType = 15 // Mail exchange
	TypeTXT   RecordType = 16 // Text strings
)
