package rison

import "fmt"

func ExampleMarshal() {
	v := struct {
		I int64       `json:"i"`
		F float64     `json:"f"`
		S string      `json:"s"`
		B bool        `json:"b"`
		P *bool       `json:"p"`
		A []int64     `json:"a"`
		X interface{} `json:"x"`
	}{
		I: 1,
		F: 2.3,
		S: "str",
		B: true,
		P: nil,
		A: []int64{7, 8, 9},
		X: map[string]interface{}{"y": "Y"},
	}
	r, _ := Marshal(&v, Mode_Rison)
	fmt.Println(string(r))
	// Output: (a:!(7,8,9),b:!t,f:2.3,i:1,p:!n,s:str,x:(y:Y))
}

func ExampleFromJSON() {
	j := `[1,2.3,"str","-ing","true","nil",{"a":"b"},[7,8,9]]`
	r, err := FromJSON([]byte(j), Mode_Rison)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(r))
	// Output: !(1,2.3,str,'-ing',true,nil,(a:b),!(7,8,9))
}
