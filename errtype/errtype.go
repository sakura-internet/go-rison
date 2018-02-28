package errtype

// ErrType is an enum type of error
type ErrType int

const (
	// Internal is an error indicating an internal error occurred.
	Internal ErrType = iota
	// Encoding is an error indicating encoding failed.
	Encoding
	// EmptyString is an error indicating the string is empty.
	EmptyString
	// UnmatchedPair is an error indicating characters such as parentheses are not paired.
	UnmatchedPair
	// MissingCharacter is an error indicating necessary characters are missing.
	MissingCharacter
	// MissingCharacterAfterEscape is an error indicating there is no character after the escape character.
	MissingCharacterAfterEscape
	// ExtraCharacter is an error indicating extra characters.
	ExtraCharacter
	// ExtraCharacterAfterRison is an error indicating there are extra characters after valid Rison.
	ExtraCharacterAfterRison
	// InvalidLiteral is an error indicating an invalid literal was found.
	InvalidLiteral
	// InvalidCharacter is an error indicating an invalid character was found.
	InvalidCharacter
	// InvalidTypeOfObjectKey is an error indicating an invalid type object key was found.
	InvalidTypeOfObjectKey
	// InvalidStringEscape is an error indicating an invalid string escape was found.
	InvalidStringEscape
	// InvalidNumber is an error indicating an invalid number was found.
	InvalidNumber
	// InvalidLargeExp is an error indicating an upper case "E" is used as an exponent.
	InvalidLargeExp
)
