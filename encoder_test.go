package rison_test

import (
	"fmt"

	"github.com/sakura-internet/go-rison/v4"
)

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

func ExampleFromJSON() {
	j := `[1,2.3,"str","-ing","true","nil",{"a":"b"},[7,8,9]]`
	r, err := rison.FromJSON([]byte(j), rison.Rison)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(r))
	// Output: !(1,2.3,str,'-ing',true,nil,(a:b),!(7,8,9))
}
