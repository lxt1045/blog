package main

import (
	"log"

	"github.com/tidwall/gjson"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile) //log.Llongfile
}

func main() {
	j := `{
		"tiered_pricing":[
			{
				"count":1000,
				"quotation":8.0
			},
			{
				"count":10000,
				"quotation":5.0
			}
		]
	}`
	m, ok := gjson.Parse(j).Value().(map[string]interface{})
	if !ok {
		log.Fatalln("error")
	}
	log.Printf("%+v", m)

	log.Printf("gjson.Get:%+v", gjson.Get(j, "tiered_pricing.0.quotation"))

	// Marshal
	// output, err := sonic.Marshal(&data)
	// // Unmarshal
	// err := sonic.Unmarshal(output, &data)
}
