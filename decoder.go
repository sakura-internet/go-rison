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

func substrLimited(str []byte, o, n int) []byte {
	if o < 0 {
		n += o
		o = 0
	}
	if n < 0 {
		n = 0
	}
	return substr(str, o, n)
}

type parser struct {
	Mode            Mode
	SkipWhitespaces bool
	string          []byte
	index           int
	buffer          *bytes.Buffer
}

type ParseError struct {
	Child error
	Err   string
	Src   []byte
	Pos   int
}

func (e *ParseError) Error() string {
	n := 5
	ll := ""
	if 0 < e.Pos-n {
		ll = ".. "
	}
	l := string(substrLimited(e.Src, e.Pos-n, n))
	c := string(substrLimited(e.Src, e.Pos, 1))
	r := string(substrLimited(e.Src, e.Pos+1, n))
	rr := ""
	if e.Pos+1+n < len(e.Src) {
		rr = " .."
	}
	w := fmt.Sprintf(` at [%d] near %s"%s" -> "%s" -> "%s"%s`, e.Pos, ll, l, c, r, rr)
	if l == "" {
		if r == "" {
			if c == "" {
				w = ""
			} else {
				w = fmt.Sprintf(` at the first character "%s"`, c)
			}
		} else {
			w = fmt.Sprintf(` at the first character "%s" -> "%s"%s`, c, r, rr)
		}
	} else if c == "" {
		w = fmt.Sprintf(` at the end of string %s"%s" -> EOS`, ll, l)
	} else if r == "" {
		w = fmt.Sprintf(` at the last character %s"%s" -> "%s"`, ll, l, c)
	}
	result := fmt.Sprintf(`%s%s`, e.Err, w)
	if e.Child != nil {
		result += "\n" + e.Child.Error()
	}
	return result
}

func (p *parser) errorf(pos int, err error, format string, args ...interface{}) error {
	i := p.index
	src := p.string
	switch p.Mode {
	case ORison:
		src = substr(src, 1, -1)
		i--
	case ARison:
		src = substr(src, 2, -1)
		i -= 2
	}
	pos += i
	return &ParseError{
		Child: err,
		Err:   fmt.Sprintf(format, args...),
		Src:   src,
		Pos:   pos,
	}
}

func (p *parser) parse(rison []byte) ([]byte, error) {
	if !utf8.Valid(rison) {
		return nil, p.errorf(0, nil, `rison must be a valid UTF-8 string`)
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
	_, err := p.readValue()
	if err != nil {
		return nil, err
	}
	j := p.buffer.Bytes()
	p.buffer = nil
	if p.index < len(p.string) {
		// TODO: error message for large E in exponential number
		c := p.string[p.index]
		return j, p.errorf(0, nil, `extra character "%c" (0x%02X) after top-level value`, c, c)
	}
	return j, nil
}

type nodeType int

const (
	nodeType_invalid nodeType = iota
	nodeType_null
	nodeType_boolean
	nodeType_number
	nodeType_string
	nodeType_array
	nodeType_object
)

func (p *parser) readValue() (nodeType, error) {
	c, ok := p.next()
	if !ok {
		return nodeType_invalid, p.errorf(0, nil, `empty expression`)
	}

	switch {
	case c == '!':
		return p.parseSpecial()
	case c == '(':
		return nodeType_object, p.parseObject()
	case c == '\'':
		return nodeType_string, p.parseQuotedString()
	case c == '-' || '0' <= c && c <= '9':
		return nodeType_number, p.parseNumber()
	}

	p.index--

	ok, err := p.parseID()
	if err != nil {
		return nodeType_invalid, err
	}
	if ok {
		return nodeType_string, nil
	}

	return nodeType_invalid, p.errorf(0, nil, `invalid character "%c" (0x%02X)`, c, c)
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
		return false, p.errorf(0, err, `id "%s" cannot be converted to JSON`, string(id))
	}
	p.index = i
	p.buffer.Write(j)
	return true, nil
}

func (p *parser) parseSpecial() (nodeType, error) {
	s := p.string
	if len(s) <= p.index {
		return nodeType_invalid, p.errorf(0, nil, `"!" at end of input`)
	}
	c := s[p.index]
	p.index++
	switch c {
	case 't':
		p.buffer.WriteString("true")
		return nodeType_boolean, nil
	case 'f':
		p.buffer.WriteString("false")
		return nodeType_boolean, nil
	case 'n':
		p.buffer.WriteString("null")
		return nodeType_null, nil
	case '(':
		return nodeType_array, p.parseArray()
	}
	return nodeType_invalid, p.errorf(-1, nil, `unknown literal "!%c"`, c)
}

func (p *parser) parseArray() error {
	notFirst := false
	p.buffer.WriteByte('[')
	for {
		c, ok := p.next()
		if !ok {
			return p.errorf(0, nil, `unmatched "!("`)
		}
		if c == ')' {
			break
		}
		if notFirst {
			if c != ',' {
				return p.errorf(-1, nil, `missing ","`)
			}
			p.buffer.WriteByte(',')
		} else if c == ',' {
			return p.errorf(-1, nil, `extra ","`)
		} else {
			p.index--
		}
		_, err := p.readValue()
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
			return p.errorf(0, nil, `unmatched "("`)
		}
		if c == ')' {
			break
		}
		if notFirst {
			if c != ',' {
				return p.errorf(-1, nil, `missing ","`)
			}
			p.buffer.WriteByte(',')
		} else if c == ',' {
			return p.errorf(-1, nil, `extra ","`)
		} else {
			p.index--
		}
		typ, err := p.readValue()
		if err != nil {
			return p.errorf(0, err, `invalid object key`)
		}
		if typ != nodeType_string {
			return p.errorf(-1, nil, `object key must be a string`)
		}
		c, ok = p.next()
		if !ok {
			return p.errorf(0, nil, `missing ":"`)
		}
		if c != ':' {
			return p.errorf(-1, nil, `missing ":" or invalid type of object key`)
		}
		p.buffer.WriteByte(':')
		_, err = p.readValue()
		if err != nil {
			return p.errorf(0, err, `invalid object value`)
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
			p.index = i
			return p.errorf(0, nil, `unmatched "'"`)
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
				p.index = i
				return p.errorf(0, nil, `invalid string escape "!%c"`, c)
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
		return p.errorf(0, err, `invalid string "%s"`, string(result))
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
		return p.errorf(0, nil, `invalid number`)
	}
	var result interface{}
	err := json.Unmarshal(t, &result)
	if err != nil {
		return p.errorf(0, err, `invalid number "%s"`, string(t))
	}
	j, err := json.Marshal(result)
	if err != nil {
		return p.errorf(0, err, `invalid number "%s"`, string(t))
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
