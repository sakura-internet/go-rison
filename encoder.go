package rison

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Marshal returns the Rison encoding of v.
//
// The object keys corresponding the struct fields can be
// specified in struct tag (not "rison" but) "json".
func Marshal(v interface{}, m Mode) ([]byte, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return FromJSON(j, m)
}

// FromJSON parses the JSON-encoded data and returns the
// Rison-encoded data that expresses the equal value.
func FromJSON(data []byte, m Mode) ([]byte, error) {
	return (&encoder{Mode: m}).encode(data, m)
}

// Encode is an alias of Marshal.
func Encode(v interface{}, m Mode) ([]byte, error) {
	return Marshal(v, m)
}

type encoder struct {
	Mode   Mode
	buffer *bytes.Buffer
}

func (e *encoder) encode(data []byte, m Mode) ([]byte, error) {
	e.buffer = bytes.NewBuffer([]byte{})

	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	vv := reflect.ValueOf(v)
	kind := vv.Kind()
	switch m {
	case ORison:
		if kind != reflect.Map {
			return nil, fmt.Errorf("only a struct or a map[string] can be encoded to the O-Rison")
		}
	case ARison:
		if !(kind == reflect.Slice || kind == reflect.Array) {
			return nil, fmt.Errorf("only a slice or an array can be encoded to the A-Rison")
		}
	}

	if bytes.Equal(data, []byte("null")) {
		return []byte("!n"), nil
	}
	if !vv.IsValid() {
		return nil, fmt.Errorf("invalid JSON: %s", string(data))
	}

	err = e.encodeValue("", vv)
	if err != nil {
		return nil, err
	}

	r := e.buffer.Bytes()
	e.buffer = nil
	n := len(r)

	switch m {
	case ORison:
		if !(3 <= n && r[0] == '(' && r[n-1] == ')') {
			return nil, fmt.Errorf("failed to encode the value to the O-Rison")
		}
		r = r[1 : n-1]
	case ARison:
		if !(4 <= n && r[0] == '!' && r[1] == '(' && r[n-1] == ')') {
			return nil, fmt.Errorf("failed to encode the value to the A-Rison")
		}
		r = r[2 : n-1]
	}

	return r, nil
}

func idOk(s string) bool {
	n := len(s)
	if n == 0 {
		return false
	}
	if 0 <= strings.IndexByte(notIDStart, s[0]) {
		return false
	}
	for i := 1; i < n; i++ {
		if 0 <= strings.IndexByte(notIDChar, s[i]) {
			return false
		}
	}
	return true
}

func (e *encoder) writeString(v reflect.Value) bool {
	if !v.CanInterface() {
		return false
	}
	s, ok := v.Interface().(string)
	if !ok {
		return false
	}
	if idOk(s) {
		e.buffer.WriteString(s)
		return true
	}
	n := len(s)
	e.buffer.WriteByte('\'')
	for i := 0; i < n; i++ {
		c := s[i]
		if c == '\'' || c == '!' {
			e.buffer.WriteByte('!')
		}
		e.buffer.WriteByte(c)
	}
	e.buffer.WriteByte('\'')
	return true
}

func (e *encoder) encodeValue(path string, v reflect.Value) error {
	var errDetail error
encodeValueError:
	switch v.Kind() {

	case reflect.Bool:
		if !v.CanInterface() {
			break
		}
		b, ok := v.Interface().(bool)
		if !ok {
			break
		}
		if b {
			e.buffer.WriteString("!t")
		} else {
			e.buffer.WriteString("!f")
		}
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if !v.CanInterface() {
			break
		}
		j, err := json.Marshal(v.Interface())
		if err != nil {
			errDetail = err
			break
		}
		j = bytes.Replace(j, []byte{'+'}, []byte{}, -1)
		e.buffer.Write(j)
		return nil

	case reflect.String:
		if e.writeString(v) {
			return nil
		}

	case reflect.Map:
		e.buffer.WriteByte('(')
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			if !keys[i].CanInterface() {
				return false
			}
			ki, ok := keys[i].Interface().(string)
			if !ok {
				return false
			}
			if !keys[j].CanInterface() {
				return true
			}
			kj, ok := keys[j].Interface().(string)
			if !ok {
				return true
			}
			return ki < kj
		})
		for i, k := range keys {
			if 0 < i {
				e.buffer.WriteByte(',')
			}
			if !e.writeString(k) {
				errDetail = fmt.Errorf(`invalid key %+v`, k)
				break encodeValueError
			}
			e.buffer.WriteByte(':')
			err := e.encodeValue(path+"."+k.Interface().(string), v.MapIndex(k))
			if err != nil {
				return err
			}
		}
		e.buffer.WriteByte(')')
		return nil

	case reflect.Slice, reflect.Array:
		e.buffer.WriteString("!(")
		for i := 0; i < v.Len(); i++ {
			if 0 < i {
				e.buffer.WriteByte(',')
			}
			err := e.encodeValue(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
			if err != nil {
				return err
			}
		}
		e.buffer.WriteByte(')')
		return nil

	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			e.buffer.WriteString("!n")
			return nil
		}
		return e.encodeValue(path, v.Elem())
	}

	if path == "" {
		path = "."
	}
	var vi interface{} = v
	if v.IsValid() && v.CanInterface() {
		vi = v.Interface()
	}
	if errDetail == nil || reflect.ValueOf(errDetail).IsNil() {
		return fmt.Errorf("non-encodable %s value at %s in %+v", v.Kind(), path, vi)
	}
	return fmt.Errorf("non-encodable %s value at %s in %+v: %s", v.Kind(), path, vi, errDetail.Error())
}
