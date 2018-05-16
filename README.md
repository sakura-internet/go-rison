# go-rison

[![CircleCI](https://circleci.com/gh/sakura-internet/go-rison/tree/master.svg?style=shield&circle-token=1e751b4de85836df4db87a736dc1e9ff208fbd12)](https://circleci.com/gh/sakura-internet/go-rison)
[![Go Report Card](https://goreportcard.com/badge/github.com/sakura-internet/go-rison)](https://goreportcard.com/report/github.com/sakura-internet/go-rison)
[![codecov.io](https://codecov.io/github/sakura-internet/go-rison/coverage.svg?branch=master)](https://codecov.io/github/sakura-internet/go-rison?branch=master)
[![Godoc](https://godoc.org/github.com/sakura-internet/go-rison?status.svg)](http://godoc.org/github.com/sakura-internet/go-rison)
[![MIT License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE)

Go port of [Rison](https://github.com/Nanonid/rison).

This page describes _Rison_, a data serialization format optimized for
compactness in URIs. Rison is a slight variation of JSON that looks vastly
superior after URI encoding. Rison still expresses exactly the same set of
data structures as JSON, so data can be translated back and forth without loss
or guesswork.

## Examples

```go
func ExampleDecode() {
	r := "(id:example,str:'string',num:100,yes:!t,nil:!n,arr:!(1,2,3))"
	v, _ := rison.Decode([]byte(r), rison.Rison)
	m := v.(map[string]interface{})
	fmt.Printf(
		"id:%v, str:%v, num:%v, yes:%v, nil:%v, arr:%v",
		m["id"], m["str"], m["num"], m["yes"], m["nil"], m["arr"],
	)
	// Output: id:example, str:string, num:100, yes:true, nil:<nil>, arr:[1 2 3]
}

// The object keys corresponding the struct fields can be
// specified in struct tag (not "rison" but) "json".
type exampleStruct struct {
	I int64       `json:"i"`
	F float64     `json:"f"`
	S string      `json:"s"`
	B bool        `json:"b"`
	P *bool       `json:"p"`
	A []int64     `json:"a"`
	X interface{} `json:"x"`
}

func ExampleUnmarshal() {
	r := "(i:1,f:2.3,s:str,b:!t,p:!n,a:!(7,8,9),x:(y:Y))"
	var v exampleStruct
	_ = rison.Unmarshal([]byte(r), &v, rison.Rison)
	fmt.Printf("%+v\n", v)
	// Output: {I:1 F:2.3 S:str B:true P:<nil> A:[7 8 9] X:map[y:Y]}
}

func ExampleMarshal() {
	v := exampleStruct{
		I: 1,
		F: 2.3,
		S: "str",
		B: true,
		P: nil,
		A: []int64{7, 8, 9},
		X: map[string]interface{}{"y": "Y"},
	}
	r, _ := rison.Marshal(&v, rison.Rison)
	fmt.Println(string(r))
	// Output: (a:!(7,8,9),b:!t,f:2.3,i:1,p:!n,s:str,x:(y:Y))
}

func ExampleToJSON() {
	r := "!(1,2.3,str,'ing',true,nil,(a:b),!(7,8,9))"
	j, _ := rison.ToJSON([]byte(r), rison.Rison)
	fmt.Printf("%s\n", string(j))
	// Output: [1,2.3,"str","ing","true","nil",{"a":"b"},[7,8,9]]
}

func ExampleQuote() {
	s := "~!*()-_.,:@$'/ \"#%&+;<=>?[\\]^`{|}"
	fmt.Println(rison.QuoteString(s))
	// Output: ~!*()-_.,:@$'/+%22%23%25%26%2B%3B%3C%3D%3E%3F%5B%5C%5D%5E%60%7B%7C%7D
}

func ExampleParseError_ErrorInLang() {
	r := "!("
	_, err := rison.ToJSON([]byte(r), rison.Rison)
	fmt.Println(err.(*rison.ParseError).ErrorInLang("en"))
	fmt.Println(err.(*rison.ParseError).ErrorInLang("ja"))
	// Output:
	// unmatched "!(" (at the end of string "!(" -> EOS)
	// "!(" が閉じていません (場所: 文字列終端: "!(" → EOS)
}
```

## Descriptions

The following descriptions are some excerpts from [the original README](https://github.com/Nanonid/rison)
and [the article](https://web.archive.org/web/20130910064110/http://mjtemplate.org/examples/rison.html):

### Differences from JSON syntax

  * no whitespace is permitted except inside quoted strings. 
  * almost all character escaping is left to the uri encoder. 
  * single-quotes are used for quoting, but quotes can and should be left off strings when the strings are simple identifiers. 
  * the `e+` exponent format is forbidden, since `+` is not safe in form values and the plain `e` format is equivalent. 
  * the `E`, `E+`, and `E` exponent formats are removed. 
  * object keys should be lexically sorted when encoding. the intent is to improve url cacheability. 
  * uri-safe tokens are used in place of the standard json tokens: 
    
    |rison token|json token|meaning      |
    |:----------|:---------|:------------|
    |`'`        |`"`       |string quote |
    |`!`        |`\`       |string escape|
    |`(...)`    |`{...}`   |object       |
    |`!(...)`   |`[...]`   |array        |
    
  * the JSON literals that look like identifiers (`true`, `false` and `null`) are represented as `!` sequences: 
    
    |rison token|json token|
    |:----------|:---------|
    |`!t`       |`true`    |
    |`!f`       |`false`   |
    |`!n`       |`null`    |

The `!` character plays two similar but different roles, as an escape
character within strings, and as a marker for special values. This may be
confusing.

Notice that services can distinguish Rison-encoded strings from JSON-encoded
strings by checking the first character. Rison structures start with `(` or
`!(`. JSON structures start with `[` or `{`. This means that a service which
expects a JSON encoded object or array can accept Rison-encoded objects
without loss of compatibility.

### Interaction with URI %-encoding

Rison syntax is designed to produce strings that be legible after being [form-
encoded](http://www.w3.org/TR/html4/interact/forms.html#form-content-type) for
the [query](http://gbiv.com/protocols/uri/rfc/rfc3986.html#query) section of a
URI. None of the characters in the Rison syntax need to be URI encoded in that
context, though the data itself may require URI encoding. Rison tries to be
orthogonal to the %-encoding process - it just defines a string format that
should survive %-encoding with very little bloat. Rison quoting is only
applied when necessary to quote characters that might otherwise be interpreted
as special syntax.

Note that most URI encoding libraries are very conservative, percent-encoding
many characters that are legal according to [RFC
3986](http://gbiv.com/protocols/uri/rfc/rfc3986.html). For example,
Javascript's builtin `encodeURIComponent()` function will still make Rison
strings difficult to read. The rison.js library includes a more tolerant URI
encoder.

Rison uses its own quoting for strings, using the single quote (`**'**`) as a
string delimiter and the exclamation point (`**!**`) as the string escape
character. Both of these characters are legal in uris. Rison quoting is
largely inspired by Unix shell command line parsing.

All Unicode characters other than `**'**` and `**!**` are legal inside quoted
strings. This includes newlines and control characters. Quoting all such
characters is left to the %-encoding process.

### Grammar

modified from the [json.org](https://web.archive.org/web/20130910064110/http://json.org/) grammar.

- _object_
  - `()`
  - `(` _members_ `)`
- _members_
  - _pair_
  - _pair_ `,` _members_
- _pair_
  - _key_ `:` _value_
- _array_
  - `!()`
  - `!(` _elements_ `)`
- _elements_
  - _value_
  - _value_ `,` _elements_
- _key_
  - _id_
  - _string_
- _value_
  - _id_
  - _string_
  - _number_
  - _object_
  - _array_
  - `!t`
  - `!f`
  - `!n`
    <br>
    　　　　────────────
- _id_
  - _idstart_
  - _idstart_ _idchars_
- _idchars_
  - _idchar_
  - _idchar_ _idchars_
- _idchar_
  - any alphanumeric ASCII character
  - any ASCII character from the set `-` `_` `.` `/` `~`
  - any non-ASCII Unicode character
- _idstart_
  - any _idchar_ not in `-`, _digit_
    <br>
    　　　　────────────
- _string_
  - `''`
  - `'` _strchars_ `'`
- _strchars_
  - _strchar_
  - _strchar_ _strchars_
- _strchar_
  - any Unicode character except ASCII `'` and `!`
  - `!!`
  - `!'`
    <br>
    　　　　────────────
- _number_
  - _int_
  - _int_ _frac_
  - _int_ _exp_
  - _int_ _frac_ _exp_
- _int_
  - _digit_
  - _digit1-9_ _digits_
  - `-` digit
  - `-` digit1-9 digits
- _frac_
  - `.` _digits_
- _exp_
  - _e_ _digits_
- _digits_
  - _digit_
  - _digit_ _digits_
- _e_
  - `e`
  - `e-`
