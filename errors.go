package rison

import "fmt"

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

// ParseError is an error type to be raised by parser
type ParseError struct {
	Child error
	Type  ErrorType
	Args  []interface{}
	Src   []byte
	Pos   int
}

func (e *ParseError) Error() string {
	return e.ErrorInLang("en")
}

// ErrorInLang returns the error message in specified language.
func (e *ParseError) ErrorInLang(lang string) string {
	desc, ok := errPosDesc[lang]
	if !ok {
		desc = errPosDesc["en"]
	}
	n := 5
	ll := ""
	if 0 < e.Pos-n {
		ll = desc[errPos_ellipsisLeft]
	}
	l := string(substrLimited(e.Src, e.Pos-n, n))
	c := string(substrLimited(e.Src, e.Pos, 1))
	r := string(substrLimited(e.Src, e.Pos+1, n))
	rr := ""
	if e.Pos+1+n < len(e.Src) {
		rr = desc[errPos_ellipsisRight]
	}
	w := fmt.Sprintf(desc[errPos_near], e.Pos, ll, l, c, r, rr)
	if l == "" {
		if r == "" {
			if c == "" {
				w = ""
			} else {
				w = fmt.Sprintf(desc[errPos_first], c)
			}
		} else {
			w = fmt.Sprintf(desc[errPos_start], c, r, rr)
		}
	} else if c == "" {
		w = fmt.Sprintf(desc[errPos_end], ll, l)
	} else if r == "" {
		w = fmt.Sprintf(desc[errPos_last], ll, l, c)
	}
	msgdef, ok := errorMessage[lang]
	if !ok {
		msgdef = errorMessage["en"]
	}
	msgfmt, ok := msgdef[e.Type]
	var msg string
	if ok {
		msg = fmt.Sprintf(msgfmt, e.Args...)
	} else {
		msg = fmt.Sprintf(msgdef[ErrorType_Internal], fmt.Sprintf("err=%d", int(e.Type)))
	}
	result := msg + w
	//if e.Child != nil {
	//	result += "\n" + e.Child.Error()
	//}
	return result
}
