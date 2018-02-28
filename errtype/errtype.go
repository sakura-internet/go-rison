package errtype

// ErrType is an enum type of error
type ErrType int

const (
	Internal ErrType = iota
	Encoding
	EmptyString
	UnmatchedPair
	MissingCharacter
	MissingCharacterAfterEscape
	ExtraCharacter
	ExtraCharacterAfterRison
	InvalidLiteral
	InvalidCharacter
	InvalidTypeOfObjectKey
	InvalidStringEscape
	InvalidNumber
	InvalidLargeExp
)
