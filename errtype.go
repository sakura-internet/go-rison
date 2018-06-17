package rison

// ErrType is an enum type of error
type ErrType int

const (
	// EInternal is an error indicating an internal error occurred.
	EInternal ErrType = iota
	// EEncoding is an error indicating encoding failed.
	EEncoding
	// EEmptyString is an error indicating the string is empty.
	EEmptyString
	// EUnmatchedPair is an error indicating characters such as parentheses are not paired.
	EUnmatchedPair
	// EMissingCharacter is an error indicating necessary characters are missing.
	EMissingCharacter
	// EMissingCharacterAfterEscape is an error indicating there is no character after the escape character.
	EMissingCharacterAfterEscape
	// EExtraCharacter is an error indicating extra characters.
	EExtraCharacter
	// EExtraCharacterAfterRison is an error indicating there are extra characters after valid Rison.
	EExtraCharacterAfterRison
	// EInvalidLiteral is an error indicating an invalid literal was found.
	EInvalidLiteral
	// EInvalidCharacter is an error indicating an invalid character was found.
	EInvalidCharacter
	// EInvalidTypeOfObjectKey is an error indicating an invalid type object key was found.
	EInvalidTypeOfObjectKey
	// EInvalidStringEscape is an error indicating an invalid string escape was found.
	EInvalidStringEscape
	// EInvalidNumber is an error indicating an invalid number was found.
	EInvalidNumber
	// EInvalidLargeExp is an error indicating an upper case "E" is used as an exponent.
	EInvalidLargeExp
)
