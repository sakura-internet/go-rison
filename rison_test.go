package rison

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var testCases = map[string]string{

	// quoted strings
	"''":                `""`,
	"'0a'":              `"0a"`,
	"'abc def'":         `"abc def"`,
	"'-h'":              `"-h"`,
	"'user@domain.com'": `"user@domain.com"`,
	"'US $10'":          `"US $10"`,
	"'wow!!'":           `"wow!"`,
	"'can!'t'":          `"can't"`,

	// bare strings
	"G.":         `"G."`,
	"a":          `"a"`,
	"a-z":        `"a-z"`,
	"domain.com": `"domain.com"`,

	// numbers
	"0":     `0`,
	"1":     `1`,
	"42":    `42`,
	"1.5":   `1.5`,
	"99.99": `99.99`,
	"-3":    `-3`,
	"-33":   `-33`,
	"1e30":  `1e+30`,
	"1e-30": `1e-30`,
	//"1E30":  `1e+30`,
	"1.5e2": `150`,

	// other primitives
	"!t": `true`,
	"!f": `false`,
	"!n": `null`,

	// objects
	"()":                            `{}`,
	"(a:0)":                         `{"a":0}`,
	"(a:0,b:1)":                     `{"a":0,"b":1}`,
	"(a:0,b:foo,c:'23skidoo')":      `{"a":0,"b":"foo","c":"23skidoo"}`,
	"(a:!n)":                        `{"a":null}`,
	"(id:!n,type:/common/document)": `{"id":null,"type":"/common/document"}`,
	`(any:json,yes:!t)`:             `{"any":"json","yes":true}`,

	// arrays
	"!()":            `[]`,
	"!(!n)":          `[null]`,
	"!(1,2,3)":       `[1,2,3]`,
	"!(foo,bar)":     `["foo","bar"]`,
	"!(!t,!f,!n,'')": `[true,false,null,""]`,

	// complex objects
	`(A:(B:(C:(D:E,F:G)),H:(I:(J:K,L:M))))`:              `{"A":{"B":{"C":{"D":"E","F":"G"}},"H":{"I":{"J":"K","L":"M"}}}}`,
	`!(A,B,(supportsObjects:!t))`:                        `["A","B",{"supportsObjects":true}]`,
	"(foo:bar,baz:!(1,12e40,0.42,(a:!t,'0':!f,'1':!n)))": `{"foo":"bar","baz":[1,12e40,0.42,{"a":true,"0":false,"1":null}]}`,

	// character codes
	"'Control-F: \u0006'":     `"Control-F: \u0006"`,
	"'Null \u0000 character'": `"Null \u0000 character"`,
	"'Unicode: ௫'":            `"Unicode: ௫"`,
	"(花:上野,柳:銀座,月:隅田)":        `{"花":"上野","柳":"銀座","月":"隅田"}`,
	"(🍣:🐟,🍛:🌶,🍔:🐂)":           `{"🍣":"🐟","🍛":"🌶","🍔":"🐂"}`,
}

var invalidDecodeCases = []interface{}{

	// objects
	"(",
	"(foo",
	"(foo:",
	"(foo:1",
	")",
	"())",
	"(,",
	"(,)",
	"(foo:1,)",
	"(,bar:2)",
	"(baz!:1)",
	"(qux:1!)",
	"(1not:'id')",
	"(!t:1)",
	"(!n:1)",

	// arrays
	"name:hoge,plan:!(1,2,3),availability~:disabled,size_gib-GE:100,size_gib~GE:1024,tags:stable,tags~:!(deprecated,dev),", // raises irrelevant error message
	"!(",
	"!(1",
	"!(1,",
	"!())",
	"!(,",
	"!(,)",
	"!(1,)",
	"!(,2)",

	// strings
	"'",
	"'abc",
	"'a!'!'",
	"'!",
	"'!x",

	// numbers
	"4abc",
	"-",
	"-h",
	"-1h",
	"--1",
	"1-",
	"-1-",
	"-1-1",
	"1e-",
	"1e-h",
	"1e-1h",
	"1e--1",
	"1e1-",
	"1e-1-",
	"1e-1-1",
	"1.5e+2",
	"1.5E2",
	"1.5E+2",
	"1.5E-2",
	"1e9999999999999999",

	// escape sequences
	"!",
	"!z",
	"!!!",
	"!tf",

	// spaces
	"   ",
	"foo bar",

	// others
	"",
	"!(!t!f)",
	"(a:!t,0:!f,1:!n)",
	[]byte{0xff, 0xfe, 0xfd},
}

var invalidEncodeCases = []interface{}{
	map[float64]int{1.0: 1},
	complex(.0, 1.0),
	make(chan struct{}),
	func() {},
}

func dumpValue(v interface{}) string {
	j, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%#v", v)
	}
	return string(j)
}

func isObjectRison(r []byte) bool {
	n := len(r)
	return 3 <= n && r[0] == '(' && r[n-1] == ')'
}

func isArrayRison(r []byte) bool {
	n := len(r)
	return 4 <= n && r[0] == '!' && r[1] == '(' && r[n-1] == ')'
}

func testModes(r []byte) []Mode {
	modes := []Mode{Rison}
	if isObjectRison(r) {
		modes = append(modes, ORison)
	}
	if isArrayRison(r) {
		modes = append(modes, ARison)
	}
	return modes
}

func mustConvertMode(r []byte, mode Mode) []byte {
	switch mode {
	case ORison:
		if !isObjectRison(r) {
			panic("must be a object")
		}
		r = r[1 : len(r)-1]
	case ARison:
		if !isArrayRison(r) {
			panic("must be an array")
		}
		r = r[2 : len(r)-1]
	}
	return r
}

func testDecodeEncodeImpl(t *testing.T, object interface{}, r, j []byte, mode Mode) {
	r = mustConvertMode(r, mode)
	rs := string(r)
	js := string(j)
	decoded, err := Decode(r, mode)
	if err != nil {
		t.Errorf("decoding %s : want %s, got error `%s`", rs, js, err.Error())
	} else if !reflect.DeepEqual(object, decoded) {
		t.Errorf("decoding %s : want %s, got %s", rs, js, dumpValue(decoded))
	}

	encoded, err := Encode(object, mode)
	if err != nil {
		t.Errorf("encoding %s : want %s, got error `%s`", js, rs, err.Error())
	} else {
		redecoded, err := Decode(encoded, mode)
		if err != nil {
			t.Errorf("encoding %s : want %s, got %s and error `%s`", js, rs, string(encoded), err.Error())
		} else if !reflect.DeepEqual(object, redecoded) {
			t.Errorf("encoding %s : want %s, got %s", js, rs, string(encoded))
		}
	}
}

func TestDecodeEncode(t *testing.T) {
	for rs, js := range testCases {
		r := []byte(rs)
		j := []byte(js)
		var object interface{}
		err := json.Unmarshal(j, &object)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range testModes(r) {
			testDecodeEncodeImpl(t, object, r, j, m)
		}
	}
}

func TestDecodeDeepNestedObject(t *testing.T) {
	l := ""
	r := ""
	for i := 0; i < 10000; i++ {
		l += "(a:1,b:"
		r += ",c:3)"
	}
	l += "2"
	_, err := Decode([]byte(l+r), Rison)
	if err != nil {
		t.Errorf("decoding %s .. : want no error, got error `%s`", l[:100], err.Error())
	}
}

func TestDecodeDeepNestedArray(t *testing.T) {
	l := ""
	r := ""
	for i := 0; i < 10000; i++ {
		l += "!(!(),"
		r += ",!())"
	}
	l += "!()"
	_, err := Decode([]byte(l+r), Rison)
	if err != nil {
		t.Errorf("decoding %s .. : want no error, got error `%s`", l[:100], err.Error())
	}
}

func indent(s string) string {
	t := "\t\t"
	return t + strings.Replace(s, "\n", "\n"+t, -1)
}

func testDecodeErrorsImpl(t *testing.T, r []byte, mode Mode) {
	r = mustConvertMode(r, mode)
	decoded, err := Decode(r, mode)
	if err == nil {
		t.Errorf("decoding %s : want *ParseError, got %s", r, dumpValue(decoded))
	}
	e, ok := err.(*ParseError)
	if !ok {
		t.Errorf("decoding %s : want *ParseError, got else", r)
	}
	fmt.Printf(`"%s"`+"\n", string(r))
	fmt.Println(indent(e.ErrorInLang("en")))
	fmt.Println(indent(e.ErrorInLang("ja")))
}

func TestDecodeErrors(t *testing.T) {
	for _, rs := range invalidDecodeCases {
		r, ok := rs.([]byte)
		if !ok {
			r = []byte(rs.(string))
		}
		for _, m := range testModes(r) {
			testDecodeErrorsImpl(t, r, m)
		}
	}
}

func TestEncodeErrors(t *testing.T) {
	for _, v := range invalidEncodeCases {
		encoded, err := Encode(v, Rison)
		if err == nil {
			t.Errorf("encoding %#v : want an error, got %s", v, string(encoded))
		} else {
			fmt.Printf("%#v\n\t\t%s\n", v, err.Error())
		}
	}
}

func TestEncodeORisonError(t *testing.T) {
	cases := []interface{}{1, "a", nil, true, []interface{}{}, [1]interface{}{nil}}
	for _, v := range cases {
		encoded, err := Encode(v, ORison)
		if err == nil {
			t.Errorf("encoding %#v : want an error, got %s", v, string(encoded))
		}
	}
}

func TestEncodeARisonError(t *testing.T) {
	cases := []interface{}{1, "a", nil, true, struct{}{}, map[string]interface{}{}}
	for _, v := range cases {
		encoded, err := Encode(v, ARison)
		if err == nil {
			t.Errorf("encoding %#v : want an error, got %s", v, string(encoded))
		}
	}
}

func TestQuoteString(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	for i := byte(0); i < 128; i++ {
		buf.WriteByte(i)
	}
	s := buf.String()
	qs := QuoteString(s)
	qb := Quote([]byte(s))
	if string(qb) != qs {
		t.Errorf("escaping %s .. : want %s, got %s", s, qs, string(qb))
	}
	u, err := url.QueryUnescape(qs)
	if err != nil {
		t.Errorf("unescaping %s .. : want %s, got error `%s`", qs, s, err.Error())
	}
	if u != s {
		t.Errorf("unescaping %s .. : want %s, got %s", qs, s, u)
	}
}

func TestFromJSONError(t *testing.T) {
	j := []byte(`[`)
	_, err := FromJSON(j, Rison)
	if err == nil {
		t.Errorf("FromJSON %s : want *ParseError, got nil", string(j))
	}

	j = []byte(`[]`)
	_, err = FromJSON(j, ORison)
	if err == nil {
		t.Errorf("FromJSON %s : want *ParseError, got nil", string(j))
	}

	j = []byte(`{}`)
	_, err = FromJSON(j, ARison)
	if err == nil {
		t.Errorf("FromJSON %s : want *ParseError, got nil", string(j))
	}
}

func TestInvalidEncodeValue(t *testing.T) {
	cases := []interface{}{
		func() {},
		uintptr(1),
		[]interface{}{
			func() {},
		},
		map[float64]interface{}{
			.1: "",
		},
	}

	e := &encoder{
		buffer: bytes.NewBuffer([]byte{}),
		Mode:   Rison,
	}
	for _, v := range cases {
		vv := reflect.ValueOf(v)
		err := e.encodeValue("", vv)
		if err == nil {
			t.Errorf("encodeValue %#v : want *ParseError, got nil", v)
		}
	}
}
