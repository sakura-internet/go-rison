package rison_test

import (
	"fmt"

	"github.com/sakura-internet/go-rison"
)

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
	_ = rison.Unmarshal([]byte(r), &v, rison.Mode_Rison)
	fmt.Printf("%+v\n", v)
	// Output: {I:1 F:2.3 S:str B:true P:<nil> A:[7 8 9] X:map[y:Y]}
}

func ExampleToJSON() {
	r := "!(1,2.3,str,'ing',true,nil,(a:b),!(7,8,9))"
	j, _ := rison.ToJSON([]byte(r), rison.Mode_Rison)
	fmt.Printf("%s\n", string(j))
	// Output: [1,2.3,"str","ing","true","nil",{"a":"b"},[7,8,9]]
}
