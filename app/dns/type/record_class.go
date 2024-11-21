package _type

// RecordClass represents DNS record classes
type RecordClass uint16

const (
	ClassIN RecordClass = 1 // Internet
	ClassCH RecordClass = 3 // CHAOS class
	ClassHS RecordClass = 4 // Hesiod
)
