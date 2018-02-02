package rison

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	NOT_IDCHAR        = ` '!:(),*@$`
	IDRX              = `[^` + NOT_IDCHAR + `0-9-][^` + NOT_IDCHAR + `]*`
	PARSER_WHITESPACE = " \t\n\r\f"
)

var nextId = regexp.MustCompile(`^` + IDRX)

type parser struct {
	SkipWhitespaces bool
	string          []byte
	index           int
}

func Decode(r []byte) (interface{}, error) {
	return (&parser{}).parse(r)
}

func DecodeObject(r []byte) (interface{}, error) {
	r = append([]byte{'('}, r...)
	r = append(r, ')')
	return (&parser{}).parse(r)
}

func DecodeArray(r []byte) (interface{}, error) {
	r = append([]byte{'!', '('}, r...)
	r = append(r, ')')
	return (&parser{}).parse(r)
}

func (p *parser) substr(o, n int) string {
	s := len(p.string)
	if s == 0 {
		return ""
	}
	l := o
	if l < 0 {
		l = 0
	}
	r := o + n
	if s < r {
		r = s
	}
	return string(p.string[l:r])
}

func (p *parser) error(offset int, format string, args ...interface{}) error {
	o := offset
	if o < 0 {
		o = p.index + offset
	}
	l := p.substr(o-5, 5)
	c := p.substr(o, 1)
	r := p.substr(o+1, 5)
	w := fmt.Sprintf(`%d near .. "%s" -> "%s" -> "%s" ..`, o, l, c, r)
	if l == "" {
		w = fmt.Sprintf(`the first character "%s" -> "%s" ..`, c, r)
	} else if c == "" {
		w = fmt.Sprintf(`the end of string "%s" -> EOS`, l)
	} else if r == "" {
		w = fmt.Sprintf(`the last character .. "%s" -> "%s"`, l, c)
	}
	return fmt.Errorf(`%s at %s`, fmt.Sprintf(format, args...), w)
}

func (p *parser) parse(str []byte) (interface{}, error) {
	p.string = str
	p.index = 0
	value, err := p.readValue()
	if err != nil {
		return nil, err
	}
	if p.index < len(p.string) {
		return value, p.error(0, `extra character "%c" after top-level value`, p.string[p.index])
	}
	return value, nil
}

func (p *parser) readValue() (interface{}, error) {
	c, ok := p.next()
	if !ok {
		return nil, p.error(0, `empty expression`)
	}

	switch {
	case c == '!':
		return p.parseSpecial()
	case c == '(':
		return p.parseObject()
	case c == '\'':
		return p.parseQuotedString()
	case c == '-' || '0' <= c && c <= '9':
		return p.parseNumber()
	}

	// fell through table, parse as an id

	s := p.string
	i := p.index - 1

	m := nextId.FindSubmatch(s[i:])
	if 0 < len(m) {
		id := m[0]
		p.index = i + len(id)
		return string(id), nil
	}

	return nil, p.error(-1, `invalid character: "%c"`, c)
}

func (p *parser) parseSpecial() (interface{}, error) {
	s := p.string
	if len(s) <= p.index {
		return nil, p.error(-1, `"!" at end of input`)
	}
	c := s[p.index]
	p.index++
	switch c {
	case 't':
		return true, nil
	case 'f':
		return false, nil
	case 'n':
		return nil, nil
	case '(':
		return p.parseArray()
	}
	return nil, p.error(-1, `unknown literal: "!%c"`, c)
}

func (p *parser) parseArray() (interface{}, error) {
	ar := []interface{}{}
	for {
		c, ok := p.next()
		if !ok {
			return nil, p.error(0, `unmatched "!("`)
		}
		if c == ')' {
			break
		}
		if 0 < len(ar) {
			if c != ',' {
				return nil, p.error(-1, `missing ","`)
			}
		} else if c == ',' {
			return nil, p.error(-1, `extra ","`)
		} else {
			p.index--
		}
		v, err := p.readValue()
		if err != nil {
			return nil, err
		}
		ar = append(ar, v)
	}
	return ar, nil
}

func (p *parser) parseObject() (interface{}, error) {
	o := map[string]interface{}{}
	for {
		c, ok := p.next()
		if !ok {
			return nil, p.error(0, `unmatched "("`)
		}
		if c == ')' {
			break
		}
		if 0 < len(o) {
			if c != ',' {
				return nil, p.error(-1, `missing ","`)
			}
		} else if c == ',' {
			return nil, p.error(-1, `extra ","`)
		} else {
			p.index--
		}
		k, err := p.readValue()
		if err != nil {
			return nil, err
		}
		ks, ok := k.(string)
		if !ok {
			return nil, p.error(-1, `object key must be a string`)
		}
		if c, ok := p.next(); !(ok && c == ':') {
			return nil, p.error(-1, `missing ":"`)
		}
		v, err := p.readValue()
		if err != nil {
			return nil, err
		}
		o[ks] = v
	}
	return o, nil
}

func (p *parser) parseQuotedString() (interface{}, error) {
	s := p.string
	i := p.index
	start := i
	result := []byte{}
	for {
		if len(s) <= i {
			return nil, p.error(0, `unmatched "'"`)
		}
		c := s[i]
		i++
		if c == '\'' {
			break
		}
		if c == '!' {
			if start < i-1 {
				result = append(result, s[start:i-1]...)
			}
			c = s[i]
			i++
			if c == '!' || c == '\'' {
				result = append(result, c)
			} else {
				return nil, p.error(-1, `invalid string escape: "!%c"`, c)
			}
			start = i
		}
	}
	if start < i-1 {
		result = append(result, s[start:i-1]...)
	}
	p.index = i
	return string(result), nil
}

type numberParserState int

const (
	numberParserState_end numberParserState = iota
	numberParserState_int
	numberParserState_frac
	numberParserState_exp
)

func (p *parser) parseNumber() (interface{}, error) {
	s := p.string
	i := p.index
	start := i - 1
	state := numberParserState_int
	permittedSigns := []byte{'-'}
	for state != numberParserState_end {
		if len(s) <= i {
			i++
			break
		}
		c := s[i]
		i++
		if '0' <= c && c <= '9' {
			continue
		}
		if 0 <= bytes.IndexByte(permittedSigns, c) {
			permittedSigns = []byte{}
			continue
		}
		switch state {
		case numberParserState_int:
			if c == '.' {
				state = numberParserState_frac
			} else if c == 'e' {
				state = numberParserState_exp
				permittedSigns = []byte{'-'}
			} else {
				state = numberParserState_end
			}
		case numberParserState_frac:
			if c == 'e' {
				state = numberParserState_exp
				permittedSigns = []byte{'-'}
			} else {
				state = numberParserState_end
			}
		default:
			state = numberParserState_end
		}
	}
	i--
	p.index = i
	t := s[start:i]
	if string(t) == "-" {
		return nil, p.error(-1, `invalid number`)
	}
	var result interface{}
	err := json.Unmarshal(t, &result)
	if err != nil {
		return nil, p.error(-1, `invalid number "%s"`, string(t))
	}
	return result, nil
}

// return the next non-whitespace character
func (p *parser) next() (byte, bool) {
	for p.index < len(p.string) {
		c := p.string[p.index]
		p.index++
		if !p.SkipWhitespaces || strings.IndexByte(PARSER_WHITESPACE, c) < 0 {
			return c, true
		}
	}
	return 0, false
}
