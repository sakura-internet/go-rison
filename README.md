# go-rison

Go port of [Rison](https://github.com/Nanonid/rison).

> This page describes _Rison_, a data serialization format optimized for
> compactness in URIs. Rison is a slight variation of JSON that looks vastly
> superior after URI encoding. Rison still expresses exactly the same set of
> data structures as JSON, so data can be translated back and forth without loss
> or guesswork.

### Differences from JSON syntax

>   * no whitespace is permitted except inside quoted strings. 
>   * almost all character escaping is left to the uri encoder. 
>   * single-quotes are used for quoting, but quotes can and should be left off strings when the strings are simple identifiers. 
>   * the `e+` exponent format is forbidden, since `+` is not safe in form values and the plain `e` format is equivalent. 
>   * the `E`, `E+`, and `E` exponent formats are removed. 
>   * object keys should be lexically sorted when encoding. the intent is to improve url cacheability. 
>   * uri-safe tokens are used in place of the standard json tokens: 
>
> rison token json token  meaning
>
> * `'` `"` string quote
> * `!` `\` string escape
> * `(...)` `{...}` object
> * `!(...)` `[...]` array
>
> * the JSON literals that look like identifiers (`true`, `false` and `null`) are represented as `!` sequences: 
>
> rison token json token
>
> * `!t` true
> * `!f` false
> * `!n` null
>
> The `!` character plays two similar but different roles, as an escape
> character within strings, and as a marker for special values. This may be
> confusing.
>
> Notice that services can distinguish Rison-encoded strings from JSON-encoded
> strings by checking the first character. Rison structures start with `(` or
> `!(`. JSON structures start with `[` or `{`. This means that a service which
> expects a JSON encoded object or array can accept Rison-encoded objects
> without loss of compatibility.

### Examples

```go
func ExampleDecode() {
	r := "(id:example,str:'string',num:100,yes:!t,nil:!n,arr:!(1,2,3))"
	v, _ := rison.Decode([]byte(r), rison.Mode_Rison)
	m := v.(map[string]interface{})
	fmt.Printf(
		"id:%v, str:%v, num:%v, yes:%v, nil:%v, arr:%v",
		m["id"], m["str"], m["num"], m["yes"], m["nil"], m["arr"],
	)
	// Output: id:example, str:string, num:100, yes:true, nil:<nil>, arr:[1 2 3]
}

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
	_ = rison.Unmarshal([]byte(r), &v, rison.Mode_Rison)
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
	r, _ := rison.Marshal(&v, rison.Mode_Rison)
	fmt.Println(string(r))
	// Output: (a:!(7,8,9),b:!t,f:2.3,i:1,p:!n,s:str,x:(y:Y))
}

func ExampleToJSON() {
	r := "!(1,2.3,str,'ing',true,nil,(a:b),!(7,8,9))"
	j, _ := rison.ToJSON([]byte(r), rison.Mode_Rison)
	fmt.Printf("%s\n", string(j))
	// Output: [1,2.3,"str","ing","true","nil",{"a":"b"},[7,8,9]]
}
```
