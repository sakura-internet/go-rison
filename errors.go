package rison

// ErrorType is an enum type of error
type ErrorType int

const (
	ErrorType_Internal ErrorType = iota
	ErrorType_Encoding
	ErrorType_EmptyString
	ErrorType_UnmatchedPair
	ErrorType_MissingCharacter
	ErrorType_MissingCharacterAfterEscape
	ErrorType_ExtraCharacter
	ErrorType_ExtraCharacterAfterRison
	ErrorType_InvalidLiteral
	ErrorType_InvalidCharacter
	ErrorType_InvalidTypeOfObjectKey
	ErrorType_InvalidStringEscape
	ErrorType_InvalidNumber
	ErrorType_InvalidLargeExp
)

var errorMessage = map[string]map[ErrorType]string{
	"en": {
		ErrorType_Internal:                    `internal error: %s`,
		ErrorType_Encoding:                    `Rison must be a valid UTF-8 string`,
		ErrorType_EmptyString:                 `empty string`,
		ErrorType_UnmatchedPair:               `unmatched "%s"`,
		ErrorType_MissingCharacter:            `missing "%c"`,
		ErrorType_MissingCharacterAfterEscape: `missing character after "!"`,
		ErrorType_ExtraCharacter:              `extra character "%c"`,
		ErrorType_ExtraCharacterAfterRison:    `extra character "%c" after valid Rison`,
		ErrorType_InvalidLiteral:              `invalid literal "!%c"`,
		ErrorType_InvalidCharacter:            `invalid character "%c"`,
		ErrorType_InvalidTypeOfObjectKey:      `object key must be a string`,
		ErrorType_InvalidStringEscape:         `invalid string escape "!%c"`,
		ErrorType_InvalidNumber:               `invalid number "%s"`,
		ErrorType_InvalidLargeExp:             `large case "E" for exponent cannot be used`,
	},
	"ja": {
		ErrorType_Internal:                    `内部エラー: %s`,
		ErrorType_Encoding:                    `Risonは正しいUTF-8文字列である必要があります`,
		ErrorType_EmptyString:                 `文字列が空です`,
		ErrorType_UnmatchedPair:               `"%s" が閉じていません`,
		ErrorType_MissingCharacter:            `"%c" が必要です`,
		ErrorType_MissingCharacterAfterEscape: `"!" の後に文字が必要です`,
		ErrorType_ExtraCharacter:              `"%c" が余分です`,
		ErrorType_ExtraCharacterAfterRison:    `正しいRisonの後に余分な文字 "%c" が見つかりました`,
		ErrorType_InvalidLiteral:              `不正なリテラル "!%c" が見つかりました`,
		ErrorType_InvalidCharacter:            `不正な文字 "%c" が見つかりました`,
		ErrorType_InvalidTypeOfObjectKey:      `オブジェクトキーは文字列である必要があります`,
		ErrorType_InvalidStringEscape:         `不正なエスケープ文字列 "!%c" が見つかりました`,
		ErrorType_InvalidNumber:               `不正な数値 "%s" が見つかりました`,
		ErrorType_InvalidLargeExp:             `指数表記に大文字の "E" は使用できません`,
	},
}

type errPos int

const (
	errPos_near errPos = iota
	errPos_first
	errPos_start
	errPos_end
	errPos_last
	errPos_ellipsisLeft
	errPos_ellipsisRight
)

var errPosDesc = map[string]map[errPos]string{
	"en": {
		errPos_near:          ` (at [%d] near %s"%s" -> "%s" -> "%s"%s)`,
		errPos_first:         ` (at the first character "%s")`,
		errPos_start:         ` (at the first character "%s" -> "%s"%s)`,
		errPos_end:           ` (at the end of string %s"%s" -> EOS)`,
		errPos_last:          ` (at the last character %s"%s" -> "%s")`,
		errPos_ellipsisLeft:  `.. `,
		errPos_ellipsisRight: ` ..`,
	},
	"ja": {
		errPos_near:          ` (場所: [%d]付近: %s"%s" → "%s" → "%s"%s)`,
		errPos_first:         ` (場所: 先頭文字: "%s")`,
		errPos_start:         ` (場所: 先頭文字付近: "%s" → "%s"%s)`,
		errPos_end:           ` (場所: 文字列終端: %s"%s" → EOS)`,
		errPos_last:          ` (場所: 終端文字: %s"%s" → "%s")`,
		errPos_ellipsisLeft:  `〜 `,
		errPos_ellipsisRight: ` 〜`,
	},
}
