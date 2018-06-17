package rison

import (
	"fmt"
)

var errorMessage = map[string]map[ErrType]string{
	"en": {
		EInternal:                    `internal error: %s`,
		EEncoding:                    `Rison must be a valid UTF-8 string`,
		EEmptyString:                 `empty string`,
		EUnmatchedPair:               `unmatched "%s"`,
		EMissingCharacter:            `missing "%c"`,
		EMissingCharacterAfterEscape: `missing character after "!"`,
		EExtraCharacter:              `extra character "%c"`,
		EExtraCharacterAfterRison:    `extra character "%c" after valid Rison`,
		EInvalidLiteral:              `invalid literal "!%c"`,
		EInvalidCharacter:            `invalid character "%c"`,
		EInvalidTypeOfObjectKey:      `object key must be a string`,
		EInvalidStringEscape:         `invalid string escape "!%c"`,
		EInvalidNumber:               `invalid number "%s"`,
		EInvalidLargeExp:             `large case "E" for exponent cannot be used`,
	},
	"ja": {
		EInternal:                    `内部エラー: %s`,
		EEncoding:                    `Risonは正しいUTF-8文字列である必要があります`,
		EEmptyString:                 `文字列が空です`,
		EUnmatchedPair:               `"%s" が閉じていません`,
		EMissingCharacter:            `"%c" が必要です`,
		EMissingCharacterAfterEscape: `"!" の後に文字が必要です`,
		EExtraCharacter:              `"%c" が余分です`,
		EExtraCharacterAfterRison:    `正しいRisonの後に余分な文字 "%c" が見つかりました`,
		EInvalidLiteral:              `不正なリテラル "!%c" が見つかりました`,
		EInvalidCharacter:            `不正な文字 "%c" が見つかりました`,
		EInvalidTypeOfObjectKey:      `オブジェクトキーは文字列である必要があります`,
		EInvalidStringEscape:         `不正なエスケープ文字列 "!%c" が見つかりました`,
		EInvalidNumber:               `不正な数値 "%s" が見つかりました`,
		EInvalidLargeExp:             `指数表記に大文字の "E" は使用できません`,
	},
}

type errPos int

const (
	errPosNear errPos = iota
	errPosFirst
	errPosStart
	errPosEnd
	errPosLast
	errPosEllipsisLeft
	errPosEllipsisRight
)

var errLangs = []string{"en", "ja"}

var errPosDesc = map[string]map[errPos]string{
	"en": {
		errPosNear:          ` (at [%d] near %s"%s" -> "%s" -> "%s"%s)`,
		errPosFirst:         ` (at the first character "%s")`,
		errPosStart:         ` (at the first character "%s" -> "%s"%s)`,
		errPosEnd:           ` (at the end of string %s"%s" -> EOS)`,
		errPosLast:          ` (at the last character %s"%s" -> "%s")`,
		errPosEllipsisLeft:  `.. `,
		errPosEllipsisRight: ` ..`,
	},
	"ja": {
		errPosNear:          ` (場所: [%d]付近: %s"%s" → "%s" → "%s"%s)`,
		errPosFirst:         ` (場所: 先頭文字: "%s")`,
		errPosStart:         ` (場所: 先頭文字付近: "%s" → "%s"%s)`,
		errPosEnd:           ` (場所: 文字列終端: %s"%s" → EOS)`,
		errPosLast:          ` (場所: 終端文字: %s"%s" → "%s")`,
		errPosEllipsisLeft:  `〜 `,
		errPosEllipsisRight: ` 〜`,
	},
}

// ParseError is an error type to be raised by parser
type ParseError struct {
	Child error
	Type  ErrType
	Args  []interface{}
	Src   []byte
	Pos   int
}

func (e *ParseError) Error() string {
	return e.ErrorInLang("en")
}

// Langs returns supported languages.
func (e *ParseError) Langs() []string {
	return errLangs
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
		ll = desc[errPosEllipsisLeft]
	}
	l := string(substrLimited(e.Src, e.Pos-n, n))
	c := string(substrLimited(e.Src, e.Pos, 1))
	r := string(substrLimited(e.Src, e.Pos+1, n))
	rr := ""
	if e.Pos+1+n < len(e.Src) {
		rr = desc[errPosEllipsisRight]
	}
	w := fmt.Sprintf(desc[errPosNear], e.Pos, ll, l, c, r, rr)
	if l == "" {
		if r == "" {
			if c == "" {
				w = ""
			} else {
				w = fmt.Sprintf(desc[errPosFirst], c)
			}
		} else {
			w = fmt.Sprintf(desc[errPosStart], c, r, rr)
		}
	} else if c == "" {
		w = fmt.Sprintf(desc[errPosEnd], ll, l)
	} else if r == "" {
		w = fmt.Sprintf(desc[errPosLast], ll, l, c)
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
		msg = fmt.Sprintf(msgdef[EInternal], fmt.Sprintf("err=%d", int(e.Type)))
	}
	result := msg + w
	//if e.Child != nil {
	//	result += "\n" + e.Child.Error()
	//}
	return result
}
