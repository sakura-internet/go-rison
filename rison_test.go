package rison

import (
	"encoding/json"
	"fmt"
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
	"(1not:'id')",

	// arrays
	"name:hoge,plan:!(1,2,3),availability~:disabled,size_gib-GE:100,size_gib~GE:1024,tags:stable,tags~:!(deprecated,dev),", // raises irrelevant error message

	// strings
	"'",
	"'abc",
	"'a!'!'",

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

func testDecodeEncodeImpl(t *testing.T, object interface{}, r, j string, mode Mode) {
	switch mode {
	case ORison:
		r = r[1 : len(r)-1]
	case ARison:
		r = r[2 : len(r)-1]
	}

	decoded, err := Decode([]byte(r), mode)
	if err != nil {
		t.Errorf("decoding %s : want %s, got error `%s`", r, j, err.Error())
	} else if !reflect.DeepEqual(object, decoded) {
		t.Errorf("decoding %s : want %s, got %s", r, j, dumpValue(decoded))
	}

	encoded, err := Encode(object, mode)
	if err != nil {
		t.Errorf("encoding %s : want %s, got error `%s`", j, r, err.Error())
	} else {
		redecoded, err := Decode(encoded, mode)
		if err != nil {
			t.Errorf("encoding %s : want %s, got %s and error `%s`", j, r, string(encoded), err.Error())
		} else if !reflect.DeepEqual(object, redecoded) {
			t.Errorf("encoding %s : want %s, got %s", j, r, string(encoded))
		}
	}
}

func TestDecodeEncode(t *testing.T) {
	for r, j := range testCases {
		var object interface{}
		err := json.Unmarshal([]byte(j), &object)
		if err != nil {
			t.Fatal(err)
		}

		modes := []Mode{Rison}
		n := len(r)
		if 3 <= n && r[0] == '(' && r[n-1] == ')' {
			modes = append(modes, ORison)
		}
		if 4 <= n && r[0] == '!' && r[1] == '(' && r[n-1] == ')' {
			modes = append(modes, ARison)
		}

		for _, m := range modes {
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

func TestDecodeErrors(t *testing.T) {
	for _, rs := range invalidDecodeCases {
		r, ok := rs.([]byte)
		if !ok {
			r = []byte(rs.(string))
		}
		decoded, err := Decode(r, Rison)
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
