package rison

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Unmarshal parses the Rison-encoded data and stores the result
// in the value pointed to by v.
//
// The object keys corresponding the struct fields can be
// specified in struct tag (not "rison" but) "json".
func Unmarshal(data []byte, v interface{}, m Mode) error {
	j, err := ToJSON(data, m)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, v)
}

// ToJSON parses the Rison-encoded data and returns the
// JSON-encoded data that expresses the equal value.
func ToJSON(data []byte, m Mode) ([]byte, error) {
	return (&parser{Mode: m}).parse(data)
}

// Decode parses the Rison-encoded data and returns the
// result as the tree of map[string]interface{}
// (or []interface{} or scalar value).
func Decode(data []byte, m Mode) (interface{}, error) {
	j, err := ToJSON(data, m)
	if err != nil {
		return nil, err
	}
	var o interface{}
	err = json.Unmarshal(j, &o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func substr(str []byte, o, n int) []byte {
	s := len(str)
	if s == 0 {
		return []byte{}
	}
	l := o
	if l < 0 {
		l = 0
	}
	r := o + n
	if n < 0 {
		r = s + n
	}
	if s < r {
		r = s
	}
	if r <= l {
		return []byte{}
	}
	return str[l:r]
}

type parser struct {
	Mode            Mode
	SkipWhitespaces bool
	string          []byte
	index           int
	buffer          *bytes.Buffer
}

func (p *parser) error(offset int, format string, args ...interface{}) error {
	i := p.index
	s := p.string
	switch p.Mode {
	case ORison:
		s = substr(s, 1, -1)
		i--
	case ARison:
		s = substr(s, 2, -1)
		i -= 2
	}
	o := offset
	if o < 0 {
		o = i + offset
	}
	l := string(substr(s, o-5, 5))
	c := string(substr(s, o, 1))
	r := string(substr(s, o+1, 5))
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

func (p *parser) parse(rison []byte) ([]byte, error) {
	if !utf8.Valid(rison) {
		return nil, p.error(0, `rison must be a valid UTF-8 string`)
	}

	switch p.Mode {
	case ORison:
		rison = append([]byte{'('}, rison...)
		rison = append(rison, ')')
	case ARison:
		rison = append([]byte{'!', '('}, rison...)
		rison = append(rison, ')')
	}
	p.string = rison
	p.index = 0
	p.buffer = bytes.NewBuffer(make([]byte, 0, len(rison)))
	err := p.readValue()
	if err != nil {
		return nil, err
	}
	j := p.buffer.Bytes()
	p.buffer = nil
	if p.index < len(p.string) {
		return j, p.error(0, `extra character "%c" after top-level value`, p.string[p.index])
	}
	return j, nil
}

func (p *parser) readValue() error {
	c, ok := p.next()
	if !ok {
		return p.error(0, `empty expression`)
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

	p.index--

	ok, err := p.parseID()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	return p.error(0, `invalid character: "%c"`, c)
}

func (p *parser) parseID() (bool, error) {
	s := p.string
	n := len(s)
	i := p.index
	if n <= i {
		return false, nil
	}
	c := s[i]
	if 0 <= strings.IndexByte(notIDStart, c) {
		return false, nil
	}
	i++
	id := []byte{c}
	for {
		if n <= i {
			break
		}
		c := s[i]
		if 0 <= strings.IndexByte(notIDChar, c) {
			break
		}
		i++
		id = append(id, c)
	}
	j, err := json.Marshal(string(id))
	if err != nil {
		return false, p.error(-1, `invalid id "%s": %s`, string(id), err.Error())
	}
	p.index = i
	p.buffer.Write(j)
	return true, nil
}

func (p *parser) parseSpecial() error {
	s := p.string
	if len(s) <= p.index {
		return p.error(-1, `"!" at end of input`)
	}
	c := s[p.index]
	p.index++
	switch c {
	case 't':
		p.buffer.WriteString("true")
		return nil
	case 'f':
		p.buffer.WriteString("false")
		return nil
	case 'n':
		p.buffer.WriteString("null")
		return nil
	case '(':
		return p.parseArray()
	}
	return p.error(-1, `unknown literal: "!%c"`, c)
}

func (p *parser) parseArray() error {
	notFirst := false
	p.buffer.WriteByte('[')
	for {
		c, ok := p.next()
		if !ok {
			return p.error(0, `unmatched "!("`)
		}
		if c == ')' {
			break
		}
		if notFirst {
			if c != ',' {
				return p.error(-1, `missing ","`)
			}
			p.buffer.WriteByte(',')
		} else if c == ',' {
			return p.error(-1, `extra ","`)
		} else {
			p.index--
		}
		err := p.readValue()
		if err != nil {
			return err
		}
		notFirst = true
	}
	p.buffer.WriteByte(']')
	return nil
}

func (p *parser) parseObject() error {
	notFirst := false
	p.buffer.WriteByte('{')
	for {
		c, ok := p.next()
		if !ok {
			return p.error(0, `unmatched "("`)
		}
		if c == ')' {
			break
		}
		if notFirst {
			if c != ',' {
				return p.error(-1, `missing ","`)
			}
			p.buffer.WriteByte(',')
		} else if c == ',' {
			return p.error(-1, `extra ","`)
		} else {
			p.index--
		}
		err := p.readValue() // @todo must be a string
		if err != nil {
			return err
		}
		if c, ok := p.next(); !(ok && c == ':') {
			return p.error(-1, `missing ":"`)
		}
		p.buffer.WriteByte(':')
		err = p.readValue()
		if err != nil {
			return err
		}
		notFirst = true
	}
	p.buffer.WriteByte('}')
	return nil
}

func (p *parser) parseQuotedString() error {
	s := p.string
	i := p.index
	start := i
	result := []byte{}
	for {
		if len(s) <= i {
			return p.error(0, `unmatched "'"`)
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
				return p.error(-1, `invalid string escape: "!%c"`, c)
			}
			start = i
		}
	}
	if start < i-1 {
		result = append(result, s[start:i-1]...)
	}
	p.index = i
	j, err := json.Marshal(string(result))
	if err != nil {
		return p.error(-1, `invalid string "%s": %s`, string(result), err.Error())
	}
	p.buffer.Write(j)
	return nil
}

type parseNumberState int

const (
	parseNumberStateEnd parseNumberState = iota
	parseNumberStateInt
	parseNumberStateFrac
	parseNumberStateExp
)

func (p *parser) parseNumber() error {
	s := p.string
	i := p.index
	start := i - 1
	state := parseNumberStateInt
	permittedSigns := []byte{'-'}
	for state != parseNumberStateEnd {
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
		case parseNumberStateInt:
			if c == '.' {
				state = parseNumberStateFrac
			} else if c == 'e' {
				state = parseNumberStateExp
				permittedSigns = []byte{'-'}
			} else {
				state = parseNumberStateEnd
			}
		case parseNumberStateFrac:
			if c == 'e' {
				state = parseNumberStateExp
				permittedSigns = []byte{'-'}
			} else {
				state = parseNumberStateEnd
			}
		default:
			state = parseNumberStateEnd
		}
	}
	i--
	p.index = i
	t := s[start:i]
	if string(t) == "-" {
		return p.error(-1, `invalid number`)
	}
	var result interface{}
	err := json.Unmarshal(t, &result)
	if err != nil {
		return p.error(-1, `invalid number "%s": %s`, string(t), err.Error())
	}
	j, err := json.Marshal(result)
	if err != nil {
		return p.error(-1, `invalid number "%s": %s`, string(t), err.Error())
	}
	p.buffer.Write(j)
	return nil
}

// return the next non-whitespace character
func (p *parser) next() (byte, bool) {
	for p.index < len(p.string) {
		c := p.string[p.index]
		p.index++
		if !p.SkipWhitespaces || strings.IndexByte(parserWhitespace, c) < 0 {
			return c, true
		}
	}
	return 0, false
}
