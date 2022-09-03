package json

import (
	"encoding/json"
	"path/filepath"
	"testing"

	lxterrs "github.com/lxt1045/errors"
	asrt "github.com/stretchr/testify/assert"
)

func TestStruct(t *testing.T) {
	getFile := func(l string) (f string) {
		_, f = filepath.Split(l)
		f = "(" + f
		return
	}
	type Anonymous struct {
		Count int `json:"count"`
		X     string
	}

	datas := []struct {
		name   string
		bs     string
		target string
		data   interface{}
	}{

		// //Map
		// {
		// 	name:   getFile(lxterrs.NewLine("map").Error()),
		// 	bs:     `{"out": 11 , "map_0": { "count":8,"y":"yyy"}}`,
		// 	target: `{"out":11,"map_0":{"count":8,"y":"yyy"}}`,
		// 	data: &struct {
		// 		Out int                    `json:"out"`
		// 		Map map[string]interface{} `json:"map_0"`
		// 	}{},
		// },

		// 匿名类型; 指针匿名类型
		{
			name:   getFile(lxterrs.NewLine("struct").Error()),
			bs:     `{"out": 11 , "count":8,"X":"xxx"}`,
			target: `{"out":11,"count":8,"X":"xxx"}`,
			data: &struct {
				Out int `json:"out"`
				Anonymous
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("struct").Error()),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int `json:"count"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("struct").Error()),
			bs:     `{"out": 11 , "struct_0": { "count":8,"slice":[1,2,3]}}`,
			target: `{"out":11,"struct_0":{"count":8,"slice":[1,2,3]}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int   `json:"count"`
					Slice []int `json:"slice"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("slice").Error()),
			bs:     `{"count":8 , "slice":[1,2,3] }`,
			target: `{"count":8,"slice":[1,2,3]}`,
			data: &struct {
				Count int   `json:"count"`
				Slice []int `json:"slice"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("bool").Error()),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count bool `json:"count"`
				False bool `json:"false_0"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("bool-ptr").Error()),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("bool-ptr-null").Error()),
			bs:     `{"count":true , "false_0":null }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("bool-ptr-empty").Error()),
			bs:     `{"count":true }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("float64").Error()),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count float64 `json:"count"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("float64-ptr").Error()),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count *float64 `json:"count"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("int-ptr").Error()),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count *int `json:"count"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("int").Error()),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count int `json:"count"`
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("string-ptr").Error()),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN *string
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("string-notag").Error()),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN string
			}{},
		},
		{
			name:   getFile(lxterrs.NewLine("string").Error()),
			bs:     `{ "ZH_CN":"chinese", "ENUS":"English", "count":8 }`,
			target: `{"ZH_CN":"chinese"}`,
			data: &struct {
				ZHCN string `json:"ZH_CN"`
			}{},
		},
	}

	for _, d := range datas[:] {
		t.Run(d.name, func(t *testing.T) {
			err := Unmarshal([]byte(d.bs), d.data)
			if err != nil {
				t.Fatalf("%s:%v\n", d.name, err)
			}
			bs, err := json.Marshal(d.data)
			if err != nil {
				t.Fatalf("%s:%v\n", d.name, err)
			}
			asrt.Equalf(t, d.target, string(bs), d.name)
		})
	}
}
