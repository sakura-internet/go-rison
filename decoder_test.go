package rison

import (
	"fmt"
)

func ExampleDecode() {
	r := "(id:example,str:'string',num:100,yes:!t,nil:!n,arr:!(1,2,3))"
	v, _ := Decode([]byte(r), Mode_Rison)
	m := v.(map[string]interface{})
	fmt.Printf(
		"id:%v, str:%v, num:%v, yes:%v, nil:%v, arr:%v",
		m["id"], m["str"], m["num"], m["yes"], m["nil"], m["arr"],
	)
	// Output: id:example, str:string, num:100, yes:true, nil:<nil>, arr:[1 2 3]
}

func ExampleUnmarshal() {
	var v struct {
		I int64       `json:"i"`
		F float64     `json:"f"`
		S string      `json:"s"`
		B bool        `json:"b"`
		P *bool       `json:"p"`
		A []int64     `json:"a"`
		X interface{} `json:"x"`
	}
	r := "(i:1,f:2.3,s:str,b:!t,p:!n,a:!(7,8,9),x:(y:Y))"
	_ = Unmarshal([]byte(r), &v, Mode_Rison)
	fmt.Printf("%+v\n", v)
	// Output: {I:1 F:2.3 S:str B:true P:<nil> A:[7 8 9] X:map[y:Y]}
}

func ExampleToJSON() {
	r := "!(1,2.3,str,'ing',true,nil,(a:b),!(7,8,9))"
	j, _ := ToJSON([]byte(r), Mode_Rison)
	fmt.Printf("%s\n", string(j))
	// Output: [1,2.3,"str","ing","true","nil",{"a":"b"},[7,8,9]]
}
